import { MarketingContainer } from '#/components/layout';

const terms = [
  [
    'Using the service',
    'Outpipe provides network tunneling and related developer tools. You are responsible for the services you expose, the traffic you send, and the credentials you use to connect.',
  ],
  [
    'Acceptable use',
    'Do not use the service for unlawful activity, abuse, open-proxy operation, scanning, credential harvesting, malware delivery, or traffic that harms other users or the network.',
  ],
  [
    'Accounts and billing',
    'Keep your account and organization details accurate. Paid plans, usage limits, renewals, cancellations, and provider charges are shown before purchase and handled through our configured billing providers.',
  ],
  [
    'Availability',
    'We work to keep the hosted service reliable, but tunnels depend on local services, networks, relays, and third-party infrastructure. The service is provided without a guarantee of uninterrupted availability.',
  ],
];

const privacy = [
  [
    'What we process',
    'We process account details from Google or GitHub sign-in, organization membership, tunnel configuration, connection metadata, usage measurements, billing identifiers, and audit events needed to operate the service.',
  ],
  [
    'Traffic and secrets',
    'Do not place secrets in tunnel names, hostnames, request paths, or request bodies. We limit operational logging and do not intentionally store tunneled payloads as a product feature.',
  ],
  [
    'Service providers',
    'We use infrastructure, payment, analytics, and transactional email providers such as Polar, Paystack, PostHog, and Zepto Mail where required to provide the product.',
  ],
  [
    'Your choices',
    'You can manage sessions, organization membership, billing, and account deletion through the dashboard or by contacting us. Retention may vary by plan and operational requirements.',
  ],
];

function LegalPage({
  title,
  intro,
  sections,
}: {
  title: string;
  intro: string;
  sections: string[][];
}) {
  return (
    <section className="pb-16 pt-28 sm:pt-32">
      <MarketingContainer className="max-w-4xl">
        <p className="text-sm text-cyan-300">Legal</p>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight sm:text-5xl">
          {title}
        </h1>
        <p className="mt-5 max-w-2xl text-lg leading-8 text-white/50">
          {intro}
        </p>
        <p className="mt-6 text-sm text-white/35">
          Last updated: July 30, 2026
        </p>
        <div className="mt-10 space-y-8 border-t border-white/10 pt-8">
          {sections.map(([heading, text]) => (
            <article key={heading}>
              <h2 className="text-xl font-semibold">{heading}</h2>
              <p className="mt-3 max-w-3xl leading-8 text-white/55">{text}</p>
            </article>
          ))}
        </div>
      </MarketingContainer>
    </section>
  );
}

export function TermsPage() {
  return (
    <LegalPage
      title="Terms of service"
      intro="The rules for using Outpipe responsibly and safely."
      sections={terms}
    />
  );
}
export function PrivacyPage() {
  return (
    <LegalPage
      title="Privacy policy"
      intro="A clear explanation of the information needed to run your tunnels and account."
      sections={privacy}
    />
  );
}
