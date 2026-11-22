/** Announced error (alert) and/or success (status) message; renders nothing when both are empty. */
export function Notice({ error, success }: { error?: string; success?: string }) {
  return (
    <>
      {error && (
        <p role="alert" className="rounded-lg border border-red-300 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </p>
      )}
      {success && (
        <p role="status" className="rounded-lg border border-green-300 bg-green-50 p-3 text-sm text-green-800">
          {success}
        </p>
      )}
    </>
  );
}
