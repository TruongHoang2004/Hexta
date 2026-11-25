package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// IReplyAIClient is the internal AI run capability (Python AIService over gRPC). Every returned field is untrusted
// candidate data; Go applies all business gates.
type IReplyAIClient interface {
	StartRun(context.Context, types.AIRunRequest) (types.AIRun, error)
	GetRun(context.Context, string, string) (types.AIRun, error)
	CancelRun(context.Context, string, string) error
}

// ReplyScriptSource is the active script read the router uses (ScriptService.ActiveScriptVersions).
type ReplyScriptSource interface {
	ActiveScriptVersions(context.Context, string) ([]types.ActiveScriptVersion, *apperrors.Error)
}

const (
	// ReplyRouterBatch bounds one scan of due routing jobs.
	ReplyRouterBatch = 20
	// replyAIPollDelay is how long a job waits before polling an accepted AI run again.
	replyAIPollDelay = 2 * time.Second
	// replyAIMaxAttempts bounds polling (~80 s) beyond the AI service's 65 s execution budget.
	replyAIMaxAttempts = 40
	// replyHistoryLimit is the AI contract's history cap.
	replyHistoryLimit = 20
	// replyAIScriptLimit is the AI contract's script snapshot cap.
	replyAIScriptLimit = 50
	// replyTextLimit is the channel text limit; longer candidates are not proposed.
	replyTextLimit = 2000
)

// ReplyRouterService routes each customer message of a BOT_ACTIVE turn (MVP 5 FR-5.2):
//
//  1. an active script whose trigger matches (MatchScript, deterministic) gives the script answer;
//  2. otherwise the AI service is asked for a candidate: an answer, a clarifying question, or a handoff;
//  3. a handoff (or an AI failure) moves the turn to WAITING_STAFF with a reason.
//
// A candidate becomes a proposal tied to the turn, its control generation and its source. In auto_send mode Go
// sends an eligible candidate (autoSendEligible: script answers and static clarifying questions, never AI-written
// answers or anything marked requires_review)
// itself through the outbound intent pipeline, but only while the turn is still BOT_ACTIVE at the same control
// generation and context revision; otherwise the job is superseded. A proposal left pending on a BOT_ACTIVE turn
// marks the turn ReviewPending (waiting for staff, no expiry). While the reply mode is off no script read or AI call
// is made and a queued job is dropped: its turn is handed to staff (reason bot_off). Jobs are durable rows processed
// after commit, so no transaction spans the script read or the AI call, and the decision (mode included) is
// rechecked under the thread lock before anything is written.
type ReplyRouterService struct {
	store   repository.ReplyRouterStore
	scripts ReplyScriptSource
	ai      IReplyAIClient
	now     func() time.Time
}

func NewReplyRouterService(store repository.ReplyRouterStore, scripts ReplyScriptSource, ai IReplyAIClient) *ReplyRouterService {
	return &ReplyRouterService{store: store, scripts: scripts, ai: ai, now: func() time.Time { return time.Now().UTC() }}
}

// RouteDue processes one batch of due routing jobs and reports how many were seen and how many reached a decision.
func (s *ReplyRouterService) RouteDue(ctx context.Context) (seen, decided int, failure *apperrors.Error) {
	refs, err := s.store.RoutingDue(ctx, s.now(), ReplyRouterBatch)
	if err != nil {
		return 0, 0, err
	}
	for _, ref := range refs {
		if ctx.Err() != nil {
			break
		}
		ok, err := s.route(ctx, ref)
		if err != nil {
			// One failed job must not block the rest; it stays due and is retried on the next run.
			if failure == nil {
				failure = err
			}
			continue
		}
		if ok {
			decided++
		}
	}
	return len(refs), decided, failure
}

func (s *ReplyRouterService) route(ctx context.Context, ref types.ProposalRef) (bool, *apperrors.Error) {
	job, err := s.store.Proposal(ctx, ref.TenantID, ref.ProposalID, false)
	if err != nil {
		return false, err
	}
	if job.Status != types.ProposalRouting {
		return false, nil
	}
	turn, err := s.store.Turn(ctx, job.TenantID, job.TurnID)
	if err != nil {
		return false, err
	}
	// A fenced job needs no script read or AI call; apply records it as superseded. While the bot is off no AI
	// call is made either; apply drops the job.
	if job.Stale(*turn) || turn.ContextRevision != job.ContextRevision {
		return s.apply(ctx, *job, nil)
	}
	settings, err := s.store.ReplySettings(ctx, job.TenantID, false)
	if err != nil {
		return false, err
	}
	if !settings.BotOn() {
		return s.apply(ctx, *job, nil)
	}
	message, err := s.store.Message(ctx, job.TenantID, job.TriggerMessageID)
	if err != nil {
		return false, err
	}
	candidate, wait, err := s.decide(ctx, *job, *message)
	if err != nil {
		return false, err
	}
	if wait {
		return false, s.scheduleNextPoll(ctx, *job)
	}
	return s.apply(ctx, *job, candidate)
}

// decide chooses the router branch. wait=true means an accepted AI run is still in progress.
func (s *ReplyRouterService) decide(ctx context.Context, job types.ReplyProposal, message types.MessagingMessage) (*types.ReplyCandidate, bool, *apperrors.Error) {
	if message.ContentType == types.ContentUnsupportedMedia || strings.TrimSpace(message.Body) == "" {
		return handoffCandidate("insufficient_information", "unsupported_media"), false, nil
	}
	versions, err := s.scripts.ActiveScriptVersions(ctx, job.TenantID)
	if err != nil {
		return nil, false, err
	}
	if version, ok := MatchScript(message.Body, versions); ok {
		return &types.ReplyCandidate{Route: types.RouteScript, Source: types.SourceScript, ScriptVersionID: version.VersionID, Text: version.Content.Answer, Reason: "script_trigger_match"}, false, nil
	}
	return s.askAI(ctx, job, message, versions)
}

func (s *ReplyRouterService) askAI(ctx context.Context, job types.ReplyProposal, message types.MessagingMessage, versions []types.ActiveScriptVersion) (*types.ReplyCandidate, bool, *apperrors.Error) {
	last := job.Attempts+1 >= replyAIMaxAttempts
	var run types.AIRun
	var callErr error
	if job.Attempts > 0 {
		run, callErr = s.ai.GetRun(ctx, job.TenantID, job.ID)
	}
	// The first attempt starts the run. A later attempt starts it again when the AI service has no such run, i.e. an
	// earlier StartRun failed before the service committed it; StartRun is idempotent on the run ID (the job ID).
	if job.Attempts == 0 || errors.Is(callErr, types.ErrAIRunNotFound) {
		request, err := s.aiRequest(ctx, job, message, versions)
		if err != nil {
			return nil, false, err
		}
		run, callErr = s.ai.StartRun(ctx, request)
	}
	switch {
	case errors.Is(callErr, types.ErrAIUnavailable):
		return handoffCandidate("ai_error", "ai_unavailable"), false, nil
	case errors.Is(callErr, types.ErrAIRunConflict):
		// The run ID is held with different input: that run belongs to another basis and is never used.
		return handoffCandidate("ai_error", "ai_run_conflict"), false, nil
	case callErr != nil && last:
		return handoffCandidate("ai_error", "ai_unreachable"), false, nil
	case callErr != nil:
		return nil, true, nil
	}
	switch run.Status {
	case types.AIRunQueued, types.AIRunRunning:
		if !last {
			return nil, true, nil
		}
		// Best effort: the durable cancel revokes publication; a provider call may already have cost money.
		_ = s.ai.CancelRun(ctx, job.TenantID, job.ID)
		return handoffCandidate("ai_error", "ai_timeout"), false, nil
	case types.AIRunSucceeded:
	default:
		return handoffCandidate("ai_error", "ai_failed"), false, nil
	}
	return aiCandidate(job, run, versions), false, nil
}

// aiCandidate turns an untrusted AI run into a router decision. Script choices must name a currently active
// version and always use Go's stored content; free text must be within the channel limit.
func aiCandidate(job types.ReplyProposal, run types.AIRun, versions []types.ActiveScriptVersion) *types.ReplyCandidate {
	if run.ContextRevision != job.ContextRevision || run.ControlGeneration != job.ControlGeneration {
		return handoffCandidate("ai_error", "ai_basis_mismatch")
	}
	text := strings.TrimSpace(run.Text)
	textOK := text != "" && utf8.ValidString(text) && utf8.RuneCountInString(text) <= replyTextLimit
	switch run.Route {
	case "script":
		for _, version := range versions {
			if version.VersionID == run.ScriptID {
				return &types.ReplyCandidate{Route: types.RouteScript, Source: types.SourceScript, ScriptVersionID: version.VersionID, AIRunID: job.ID, Text: version.Content.Answer, Reason: "ai_selected_script"}
			}
		}
	case "generate":
		// AI-written text (a knowledge-grounded or legacy fact answer) always waits for staff review, whatever the
		// run's own labels say; its cited sources are kept for the reviewer.
		if textOK {
			return &types.ReplyCandidate{Route: types.RouteAnswer, Source: types.SourceAI, AIRunID: job.ID, Text: text, Reason: safeReason(run.Reason),
				CitedSourceIDs: safeCitations(run.CitedSourceIDs), RequiresReview: true}
		}
	case "clarify":
		// A model-written clarifying question carries requires_review (or candidate_kind generated) and is then never
		// sent automatically; the AI service's static clarifying question may be.
		if textOK {
			return &types.ReplyCandidate{Route: types.RouteClarify, Source: types.SourceAI, AIRunID: job.ID, Text: text, Reason: safeReason(run.Reason),
				CitedSourceIDs: safeCitations(run.CitedSourceIDs), RequiresReview: run.RequiresReview || run.CandidateKind == "generated"}
		}
	case "human":
		// The AI may suggest one of its handoff categories; anything else reads as insufficient information.
		reason := "insufficient_information"
		if aiHandoffReasons[run.HandoffReason] {
			reason = run.HandoffReason
		}
		candidate := handoffCandidate(reason, safeReason(run.Reason))
		candidate.AIRunID = job.ID
		return candidate
	}
	candidate := handoffCandidate("insufficient_information", safeReason(run.Reason))
	candidate.AIRunID = job.ID
	return candidate
}

// aiHandoffReasons are the handoff categories the AI service may propose (a subset of the turn's
// needs_human_reason check: staff_handoff and bot_off are Go-owned).
var aiHandoffReasons = map[string]bool{"customer_requested_staff": true, "insufficient_information": true, "ai_error": true, "business_review": true}

// replyCitationLimit is the AI contract's cited source cap.
const replyCitationLimit = 20

// safeCitations keeps well-formed, distinct cited source IDs (kind:id with a known kind and an identifier-only id).
// Citations only label a proposal for review; they never authorize anything.
func safeCitations(values []string) []string {
	var result []string
	seen := map[string]bool{}
	for _, value := range values {
		kind, id, ok := strings.Cut(value, ":")
		if !ok || seen[value] || len(result) == replyCitationLimit || !citationKinds[kind] || id == "" || len(id) > 100 {
			continue
		}
		valid := true
		for _, r := range id {
			if !(r == '_' || r == '-' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
				valid = false
				break
			}
		}
		if valid {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

var citationKinds = map[string]bool{types.KnowledgeSourceProduct: true, types.KnowledgeSourceDocument: true, types.KnowledgeSourceScript: true, "fact": true}

func handoffCandidate(handoffReason, reason string) *types.ReplyCandidate {
	return &types.ReplyCandidate{Route: types.RouteHandoff, HandoffReason: handoffReason, Reason: reason}
}

// safeReason keeps an AI decision category only when it is a short identifier.
func safeReason(value string) string {
	if value == "" || len(value) > 64 {
		return "ai_decision"
	}
	for _, r := range value {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return "ai_decision"
		}
	}
	return value
}

func (s *ReplyRouterService) aiRequest(ctx context.Context, job types.ReplyProposal, message types.MessagingMessage, versions []types.ActiveScriptVersion) (types.AIRunRequest, *apperrors.Error) {
	previous, err := s.store.RecentMessages(ctx, job.TenantID, job.TurnID, message.ID, replyHistoryLimit)
	if err != nil {
		return types.AIRunRequest{}, err
	}
	history := make([]types.AIHistoryItem, 0, len(previous))
	for _, item := range previous {
		text := []rune(strings.TrimSpace(item.Body))
		if len(text) == 0 {
			continue
		}
		if len(text) > replyTextLimit {
			text = text[:replyTextLimit]
		}
		role := "assistant"
		if item.Direction == "inbound" {
			role = "user"
		}
		history = append(history, types.AIHistoryItem{Role: role, Text: string(text)})
	}
	// The active scripts are the authorized snapshot for the OPEN model-based script evaluation: the AI service
	// may pick one (route script), and Go then uses that version's stored content.
	scripts := make([]types.AIScript, 0, min(len(versions), replyAIScriptLimit))
	for _, version := range versions {
		if len(scripts) == replyAIScriptLimit {
			break
		}
		trigger := []rune(strings.Join(version.Content.Triggers, ", "))
		if len(trigger) > 200 {
			trigger = trigger[:200]
		}
		scripts = append(scripts, types.AIScript{ID: version.VersionID, Trigger: string(trigger), Text: version.Content.Answer})
	}
	body := []rune(message.Body)
	if len(body) > 8000 {
		body = body[:8000]
	}
	// Generation and knowledge answers are opted in: the AI service answers only from the tenant knowledge Go pushed
	// (products, active scripts, knowledge documents) and every such answer comes back requiring review, so it is
	// always a pending proposal for staff (autoSendEligible).
	return types.AIRunRequest{TenantID: job.TenantID, RunID: job.ID, TurnID: job.TurnID, ContextRevision: job.ContextRevision,
		ControlGeneration: job.ControlGeneration, Message: string(body), Scripts: scripts, History: history, AllowGeneration: true,
		AllowKnowledgeAnswer: true}, nil
}

// scheduleNextPoll keeps a job whose AI run is still in progress and schedules the next poll.
func (s *ReplyRouterService) scheduleNextPoll(ctx context.Context, job types.ReplyProposal) *apperrors.Error {
	return s.store.Within(ctx, func(r repository.MessagingStore) *apperrors.Error {
		current, err := r.Proposal(ctx, job.TenantID, job.ID, true)
		if err != nil || current.Status != types.ProposalRouting || current.Version != job.Version {
			return err
		}
		now := s.now()
		current.Attempts++
		current.AvailableAt, current.UpdatedAt = now.Add(replyAIPollDelay), now
		current.Version++
		return r.SaveProposal(ctx, *current)
	})
}

// autoSendEligible is the auto_send gate for a candidate (owner decision D2, 2026-11-26): only a script answer (a
// trigger match or an AI-selected script; apply also requires the version to still be active) and an AI clarifying
// question may be sent automatically. An AI-written answer always becomes a pending proposal for staff review, in
// every mode. Whether a newly activated script needs a test step first (PRD Q5.2) is OPEN; add that rule here.
func autoSendEligible(candidate types.ReplyCandidate) bool {
	if candidate.RequiresReview {
		// Model-authored text (a generated answer or a model-written clarifying question) is never sent automatically.
		return false
	}
	return candidate.Route == types.RouteScript || candidate.Route == types.RouteClarify
}

// apply records the decision under the thread lock after rechecking the job, the turn and the reply mode. A nil
// candidate means the job was fenced or the bot was off when it was picked up.
func (s *ReplyRouterService) apply(ctx context.Context, job types.ReplyProposal, candidate *types.ReplyCandidate) (bool, *apperrors.Error) {
	decided := false
	err := s.store.Within(ctx, func(r repository.MessagingStore) *apperrors.Error {
		thread, err := r.Thread(ctx, job.TenantID, job.ThreadID, true)
		if err != nil {
			return err
		}
		p, err := r.Proposal(ctx, job.TenantID, job.ID, true)
		if err != nil {
			return err
		}
		if p.Status != types.ProposalRouting || p.Version != job.Version {
			return nil
		}
		turn, err := r.Turn(ctx, job.TenantID, job.TurnID)
		if err != nil {
			return err
		}
		settings, err := r.ReplySettings(ctx, job.TenantID, false)
		if err != nil {
			return err
		}
		now := s.now()
		p.Version++
		p.UpdatedAt = now
		person := p.RequestedBy != ""
		state := turn.HandlingState()
		fenced := p.Stale(*turn) || turn.ContextRevision != p.ContextRevision
		botOff := !settings.BotOn()
		if candidate == nil && !fenced && !botOff {
			// route() saw the bot off and made no decision, but the owner turned it on again before this commit: the
			// job is still current, so it is requeued for an immediate fresh decision instead of being dropped.
			p.AvailableAt = now
			return r.SaveProposal(ctx, *p)
		}
		decided = true
		if candidate != nil {
			p.Route, p.Source, p.ScriptVersionID, p.AIRunID, p.Text, p.Reason = candidate.Route, candidate.Source, candidate.ScriptVersionID, candidate.AIRunID, candidate.Text, candidate.Reason
			p.CitedSourceIDs = candidate.CitedSourceIDs
		}
		switch {
		case fenced:
			// Control or context changed while the job waited: the output belongs to an old basis.
			supersede(p, "")
		case botOff && state == "BOT_ACTIVE" && !person:
			// The owner turned the bot off after this customer message was queued: drop the job and hand the
			// unanswered message to staff, as the next customer message would.
			supersede(p, types.HandoffReasonBotOff)
			p.Status, p.Route = types.ProposalHandedOff, types.RouteHandoff
			if err := botHandoff(ctx, r, turn, types.HandoffReasonBotOff, map[string]any{"router_reason": types.HandoffReasonBotOff, "proposal_id": p.ID}, now); err != nil {
				return err
			}
		case botOff:
			// A person's request (or a turn no longer under the bot) while the bot is off: nothing is proposed.
			supersede(p, types.HandoffReasonBotOff)
		case candidate.Route == types.RouteHandoff && state == "BOT_ACTIVE" && !person:
			p.Status = types.ProposalHandedOff
			if err := botHandoff(ctx, r, turn, candidate.HandoffReason, map[string]any{"router_reason": candidate.Reason, "proposal_id": p.ID}, now); err != nil {
				return err
			}
		case candidate.Route == types.RouteHandoff:
			// A person asked for a proposal (staff already involved): nothing to propose, control unchanged.
			p.Status = types.ProposalNoCandidate
		case state == "BOT_ACTIVE" || (state == "STAFF_ACTIVE" && person):
			p.Status = types.ProposalPending
			if state == "BOT_ACTIVE" && !person && settings.Mode == types.ReplyModeAutoSend && autoSendEligible(*candidate) {
				active := true
				if p.Source == types.SourceScript {
					// The script may have been deactivated or archived after the router read it; only a version
					// that is still active may be sent automatically.
					if active, err = r.ScriptVersionActive(ctx, job.TenantID, p.ScriptVersionID); err != nil {
						return err
					}
				}
				if !active {
					// The source answer is no longer authorized: the proposal is never sendable, and the customer's
					// message would otherwise sit unanswered on a bot-handled turn nobody watches, so hand it to
					// staff (the same reason as an AI-selected script that no longer exists).
					supersede(p, "script_inactive")
					p.Status, p.Route = types.ProposalHandedOff, types.RouteHandoff
					if err := botHandoff(ctx, r, turn, "insufficient_information", map[string]any{"router_reason": "script_inactive", "proposal_id": p.ID}, now); err != nil {
						return err
					}
				} else {
					sent, reason, err := s.autoSend(ctx, r, *thread, turn, *p, now)
					if err != nil {
						return err
					}
					if sent != nil {
						p.Status, p.SentMessageID, p.DecidedBy = types.ProposalAutoSent, sent.ID, types.BotActorID
					} else {
						p.Reason = reason
					}
				}
			}
			// A proposal left pending under bot handling waits for staff review (suggest safety): the turn joins
			// the waiting queue and does not expire until the proposal is sent, discarded or superseded.
			if p.Status == types.ProposalPending && state == "BOT_ACTIVE" && !turn.ReviewPending {
				turn.ReviewPending = true
				turn.Version++
				if err := r.SaveTurn(ctx, *turn, false); err != nil {
					return err
				}
			}
		default:
			supersede(p, "")
		}
		return r.SaveProposal(ctx, *p)
	})
	if err != nil {
		return false, err
	}
	return decided, nil
}

// supersede records a job or proposal as never sendable, keeping no candidate text; reason replaces the decision
// reason when set.
func supersede(p *types.ReplyProposal, reason string) {
	p.Status = types.ProposalSuperseded
	p.Text, p.Source, p.ScriptVersionID, p.CitedSourceIDs = "", "", "", nil
	if reason != "" {
		p.Reason = reason
	}
}

// autoSend sends a proposal as the bot. The caller holds the thread lock and has verified BOT_ACTIVE at the
// proposal's generation and revision. A real-channel conversation whose connection is not active is left as a
// pending proposal (reason returned) instead of failing.
func (s *ReplyRouterService) autoSend(ctx context.Context, r repository.MessagingStore, thread types.MessagingThread, turn *types.MessagingTurn, p types.ReplyProposal, now time.Time) (*types.MessagingMessage, string, *apperrors.Error) {
	hash := messagingHash(struct{ Actor, Turn, Proposal, Text string }{types.BotActorID, turn.ID, p.ID, p.Text})
	message, intentID, reason, err := recordProposalReply(ctx, r, thread, turn, types.BotActorID, p.Text, p.ID, hash, now)
	if err != nil || message == nil {
		return nil, reason, err
	}
	if err := r.Audit(ctx, turn.TenantID, types.BotActorID, "messaging.bot_auto_send", "messaging_turn", turn.ID, map[string]any{
		"proposal_id": p.ID, "message_id": message.ID, "intent_id": intentID, "route": p.Route, "source": p.Source,
		"script_version_id": p.ScriptVersionID, "ai_run_id": p.AIRunID, "control_generation": turn.ControlGeneration,
	}); err != nil {
		return nil, "", err
	}
	return message, "", nil
}

// recordProposalReply writes a proposal's reply on the turn: a real channel thread gets a pending delivery intent
// (activity counts only once the connector confirms it), a simulator thread a simulated message (a confirmed send).
// It advances the turn's version and context revision and saves the turn and thread. message is nil with a reason
// when the channel connection cannot send.
func recordProposalReply(ctx context.Context, r repository.MessagingStore, thread types.MessagingThread, turn *types.MessagingTurn, actor, text, key, hash string, now time.Time) (message *types.MessagingMessage, intentID, reason string, failure *apperrors.Error) {
	if thread.ConnectionID != "" {
		conn, err := r.SendConnection(ctx, thread.TenantID, thread.ConnectionID)
		if err != nil {
			return nil, "", "", err
		}
		if conn.Status != types.ChannelStatusActive || !conn.Enabled {
			return nil, "", "channel_unavailable", nil
		}
		message, intentID, failure = enqueueChannelReply(ctx, r, *conn, thread, turn.ID, actor, text, key, hash, now)
		if failure != nil {
			return nil, "", "", failure
		}
	} else {
		message = &types.MessagingMessage{ID: messagingID(), TenantID: thread.TenantID, ThreadID: thread.ID, TurnID: turn.ID, Direction: "outbound", ActorID: actor, Body: text, RequestKey: key, RequestHash: hash, Status: "simulated", CreatedAt: now}
		if err := r.InsertMessage(ctx, *message); err != nil {
			return nil, "", "", err
		}
		turn.LastActivityAt = now
	}
	turn.Version++
	turn.ContextRevision++
	if err := r.SaveTurn(ctx, *turn, false); err != nil {
		return nil, "", "", err
	}
	thread.UpdatedAt = now
	if err := r.SaveThread(ctx, thread); err != nil {
		return nil, "", "", err
	}
	return message, intentID, "", nil
}
