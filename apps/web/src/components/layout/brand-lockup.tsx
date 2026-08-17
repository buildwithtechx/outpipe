import { Link } from '@tanstack/react-router';

type BrandLockupProps = {
  className?: string;
  iconClassName?: string;
  nameClassName?: string;
  onClick?: () => void;
};

export function BrandLockup({
  className = '',
  iconClassName = 'size-10 rounded-xl',
  nameClassName = 'font-semibold tracking-tight',
  onClick,
}: BrandLockupProps) {
  return (
    <Link
      to="/"
      className={`inline-flex items-center gap-3 ${className}`}
      onClick={onClick}
    >
      <img src="/favicon.svg" alt="" className={iconClassName} />
      <span className={nameClassName}>Outpipe</span>
    </Link>
  );
}
