/** Announced error (alert) and/or success (status) message; renders nothing when both are empty. */
export function Notice({ error, success }: { error?: string; success?: string }) {
  return (
    <>
      {error && (
        <p role="alert" className="rounded-lg border border-error bg-error/10 p-3 text-sm text-foreground">
          {error}
        </p>
      )}
      {success && (
        <p role="status" className="rounded-lg border border-success bg-success/10 p-3 text-sm text-foreground">
          {success}
        </p>
      )}
    </>
  );
}
