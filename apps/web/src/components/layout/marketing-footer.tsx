import { Link } from '@tanstack/react-router';
import {
  SiExpress,
  SiNestjs,
  SiNextdotjs,
  SiReact,
  SiVite,
} from 'react-icons/si';
import { MarketingContainer } from './marketing-container';

export function MarketingFooter() {
  return (
    <footer className="border-t border-white/10 bg-black py-14 text-white">
      <MarketingContainer>
        <div className="grid gap-10 md:grid-cols-6">
          <div className="md:col-span-2">
            <Link to="/" className="flex items-center gap-3">
              <img src="/favicon.svg" alt="" className="size-9 rounded-xl" />
              <span className="font-semibold">Outpipe</span>
            </Link>
            <p className="mt-4 max-w-xs text-sm leading-6 text-white/45">
              Secure public access for local services, previews, webhooks, and
              private networks.
            </p>
            <p className="mt-7 text-xs text-white/35">
              © {new Date().getFullYear()} TechX Innovations Limited. All rights
              reserved.
            </p>
          </div>
          <FooterGroup title="Product">
            <FooterLink to="/pricing">Pricing</FooterLink>
            <FooterLink to="/changelog">Changelog</FooterLink>
            <FooterLink to="/contact">Contact</FooterLink>
            <FooterLink to="/report-bug">Report a bug</FooterLink>
          </FooterGroup>
          <FooterGroup title="Developers">
            <Link
              to="/docs/$"
              params={{ _splat: '' }}
              className="text-sm text-white/45 transition-colors hover:text-indigo-300"
            >
              Documentation
            </Link>
            <FooterLink to="/plugins">Plugins</FooterLink>
            <Link
              to="/docs/$"
              params={{ _splat: 'cli' }}
              className="text-sm text-white/45 transition-colors hover:text-indigo-300"
            >
              CLI reference
            </Link>
          </FooterGroup>
          <FooterGroup title="Integrations">
            <FooterPlugin to="/plugins/react" icon={SiReact}>
              React
            </FooterPlugin>
            <FooterPlugin to="/plugins/vite" icon={SiVite}>
              Vite
            </FooterPlugin>
            <FooterPlugin to="/plugins/next" icon={SiNextdotjs}>
              Next.js
            </FooterPlugin>
            <FooterPlugin to="/plugins/nest" icon={SiNestjs}>
              NestJS
            </FooterPlugin>
            <FooterPlugin to="/plugins/express" icon={SiExpress}>
              Express
            </FooterPlugin>
          </FooterGroup>
          <FooterGroup title="Legal">
            <FooterLink to="/terms">Terms</FooterLink>
            <FooterLink to="/privacy">Privacy</FooterLink>
          </FooterGroup>
        </div>
      </MarketingContainer>
    </footer>
  );
}

function FooterGroup({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <h2 className="mb-4 text-sm font-semibold text-white">{title}</h2>
      <div className="flex flex-col items-start gap-3">{children}</div>
    </div>
  );
}

function FooterLink({ children, ...props }: React.ComponentProps<typeof Link>) {
  return (
    <Link
      {...props}
      className="text-sm text-white/45 transition-colors hover:text-indigo-300"
    >
      {children}
    </Link>
  );
}

function FooterPlugin({
  children,
  icon: Icon,
  ...props
}: Omit<React.ComponentProps<typeof Link>, 'children'> & {
  children: React.ReactNode;
  icon: React.ComponentType<{ className?: string }>;
}) {
  return (
    <Link
      {...props}
      className="inline-flex items-center gap-2 text-sm text-white/45 transition-colors hover:text-indigo-300"
    >
      <Icon className="size-3.5" />
      {children}
    </Link>
  );
}
