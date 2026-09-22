import { ArrowUpRight, BadgeCheck, Scale } from 'lucide-react';
import { SiGithub } from 'react-icons/si';
import { MarketingContainer } from '#/components/layout';
import { githubRepositoryUrl } from '#/lib/github';

const commitments = [
  {
    title: 'Inspectable by default',
    text: 'Read the relay, clients, protocol schema, and integration code.',
    icon: SiGithub,
  },
  {
    title: 'AGPLv3 protected',
    text: 'Network-facing improvements remain available to the community.',
    icon: Scale,
  },
  {
    title: 'Public protocol',
    text: 'Build clients and adapters against a versioned contract.',
    icon: BadgeCheck,
  },
];

export function OpenSourceSection() {
  return (
    <section className="pb-20 pt-12 sm:pb-24 sm:pt-16">
      <MarketingContainer>
        <div className="rounded-3xl bg-white/[0.035] p-7 sm:p-10 lg:grid lg:grid-cols-[0.85fr_1.15fr] lg:gap-16">
          <div>
            <p className="text-sm font-medium text-indigo-300">Open source</p>
            <h2 className="mt-4 text-4xl font-semibold tracking-[-0.05em] sm:text-5xl">
              Built in public.
            </h2>
            <p className="mt-5 max-w-lg text-lg leading-8 text-white/50">
              The parts that carry your traffic should be understandable. Read
              the code, follow the protocol, and contribute improvements.
            </p>
            <a
              href={githubRepositoryUrl}
              target="_blank"
              rel="noreferrer"
              className="mt-8 inline-flex items-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-semibold text-black transition-transform hover:-translate-y-0.5"
            >
              <SiGithub className="size-4" />
              Explore the repository <ArrowUpRight className="size-4" />
            </a>
          </div>
          <div className="mt-10 grid gap-3 lg:mt-0">
            {commitments.map(({ title, text, icon: Icon }) => (
              <div
                key={title}
                className="flex gap-4 rounded-2xl bg-black/25 p-5"
              >
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-indigo-300/10 text-indigo-300">
                  <Icon className="size-5" />
                </span>
                <div>
                  <h3 className="font-medium text-white">{title}</h3>
                  <p className="mt-1 text-sm leading-6 text-white/45">{text}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </MarketingContainer>
    </section>
  );
}
