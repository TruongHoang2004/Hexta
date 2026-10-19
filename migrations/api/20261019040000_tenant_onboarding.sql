CREATE TABLE public.platform_operators (
 user_id varchar(255) PRIMARY KEY,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE public.tenant_registration_requests (
 id varchar(36) PRIMARY KEY,
 applicant_id varchar(255) NOT NULL,
 name varchar(100) NOT NULL,
 slug varchar(100) NOT NULL,
 status varchar(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
 tenant_id varchar(36) REFERENCES public.tenants(id),
 reviewed_by varchar(255) NOT NULL DEFAULT '',
 review_reason text NOT NULL DEFAULT '',
 reviewed_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_tenant_registration_requests_applicant_id ON public.tenant_registration_requests(applicant_id);
CREATE INDEX idx_tenant_registration_requests_status ON public.tenant_registration_requests(status);
CREATE UNIQUE INDEX idx_tenant_registration_active_slug ON public.tenant_registration_requests(slug) WHERE status <> 'rejected';
CREATE TABLE public.tenant_invitations (
 id varchar(36) PRIMARY KEY,
 tenant_id varchar(36) NOT NULL REFERENCES public.tenants(id),
 email varchar(255) NOT NULL,
 role varchar(50) NOT NULL DEFAULT 'member' CHECK (role = 'member'),
 invited_by varchar(255) NOT NULL,
 token_hash varchar(64) NOT NULL UNIQUE,
 status varchar(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','declined','revoked')),
 expires_at timestamptz NOT NULL,
 accepted_by varchar(255) NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_tenant_invitations_tenant_id ON public.tenant_invitations(tenant_id);
CREATE UNIQUE INDEX idx_tenant_pending_invitation_email ON public.tenant_invitations(tenant_id,email) WHERE status = 'pending';
