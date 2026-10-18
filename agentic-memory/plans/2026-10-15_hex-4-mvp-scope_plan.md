# HEX-4 MVP Scope Planning

Status: Proposed scope for discussion, not approved features.

## Overview and Goal

Agree on the smallest useful Hexta release covering employee responsibility across customer communication and packing, plus AI-assisted product replies reviewed by sales. HEX-4 is complete when included, deferred and excluded capabilities are agreed. It does not require completing a document, implementing features or waiting for HEX-41/42.

## Current State Analysis

HEX-1 and HEX-5 were closed by the user. Requirements come from owner-provided descriptions, not direct field measurements. Sales uses Pancake product-specific quick-reply commands; orders are recorded in Excel and handed off verbally. The owner assigns employees. Responsibility moves to packing at handoff. Direct and acceptance-based handoffs were discussed; final MVP scope is open. Payments are excluded. Cancellation/returns remain proposed workflows. No implementation readiness or existing code coverage has been audited in this planning task.

## Task Breakdown

1. Agree on the release target: thesis demonstrator, supervised pilot or production use. Confirm available development time and team capacity before claiming feasibility.
2. Review the proposed capability matrix below. Each included capability needs an actor, a concrete completed action and a boundary.
3. Set workflow boundaries: customer/product/order minimum fields, assignment, handoff modes and completion events. Keep implementation-level details for later design tickets.
4. Choose the dashboard minimum and whether manual campaign/cost entry belongs in this release. Avoid promising ad attribution without data.
5. Define cancellation/return minimum separately from payment processing; select whole-order versus partial return handling explicitly.
6. Define the integration-independent path: manually entered example messages/product context and manual operational statuses. This is a proposed fallback, not approved by the user and not equivalent to a live integration.
7. Record agreed scope in HEX-4 and update existing direction/requirements pages. Split subsequent implementation work into independently verifiable tickets; avoid creating tickets merely to finish shared documents.

## Proposed Capability Matrix

| Capability | Proposal | Boundary requiring review |
| --- | --- | --- |
| Organization and employee access | Include | Owner, sales and packing roles; tenant isolation |
| Product and customer records | Include minimal version | Product facts/media links; customer/contact and sales owner |
| Orders and responsibility | Include | Owner assignment; sales-to-packing handoff; status history |
| Handoff modes | Decide | One configured mode first or both; owner assignment remains authoritative |
| AI product replies | Include narrow version | Product-specific draft; review/edit; no automatic sending |
| Operational dashboard | Include minimal version | Assignees, pending work, order states, sold/returned units |
| Cancellation and return tracking | Decide minimum | Operational cases, reasons, receipts; no refunds |
| Ad-effectiveness dashboard | Decide minimum | Manual source/campaign/cost data or defer; no unsupported attribution |
| Configurable lead signals | Decide | Simple explicit rules or defer; no predicted purchase probability |
| Pancake integration | Separate investigation | HEX-41 reports feasibility and updates documents; no scope block |
| Ten-product dataset | Separate preparation | HEX-42 supplies data and updates documents; no scope block |
| Autonomous replies, advanced recommendation, video understanding | Defer proposal | Preserve as future direction |
| Payroll, attendance, full HR, accounting, payment execution | Exclude proposal | User explicitly excluded transaction processing; other boundaries await scope approval |
| Full inventory, carrier/advertising integrations | Defer proposal | No implied implementation commitment |

## Risks and Edge Cases

The AI study must not replace the employee-management product goal. Simulated evaluation cannot establish field effectiveness. Return quantities and delivered units must remain distinct. Assignment, handoff and shipment statuses must not be conflated. Unknown integration access needs an explicit fallback or a separately scoped conditional feature. Do not label all proposed features as agreed.

## Definition of Done

- User agrees to included, deferred and excluded capability lists, with explicit boundaries.
- Included capabilities have observable acceptance outcomes and unresolved implementation details are recorded separately.
- Release target and capacity constraints are stated, or feasibility remains explicitly provisional.
- HEX-41/42 remain independent and own updates from their results; closed HEX-1/5 need not reopen.
- Existing Notion documents are updated with scope decisions. No code changes or automated tests are required for this scope-definition ticket.

## Artifacts

This plan; HEX-4 ticket content; subsequent targeted updates to the existing central direction and HEX-1 reference pages after decisions. No new product-document tree is required.
