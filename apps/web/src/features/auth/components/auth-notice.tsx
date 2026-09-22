type AuthNoticeProps = {
  children: string;
};

export function AuthNotice({ children }: AuthNoticeProps) {
  return (
    <p
      role="alert"
      className="mb-4 rounded-xl bg-rose-300/10 px-4 py-3 text-sm text-rose-100"
    >
      {children}
    </p>
  );
}
