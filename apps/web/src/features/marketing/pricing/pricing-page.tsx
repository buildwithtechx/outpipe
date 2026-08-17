import { Link } from '@tanstack/react-router';
import { Check, Sparkles, X } from 'lucide-react';
import { useState } from 'react';
import { MarketingContainer } from '#/components/layout';

const plans = [
  {
    name: 'Free',
    description: 'Explore the core tunnel workflow.',
    usd: 0,
    ngn: 0,
    featured: false,
    features: [
      ['2 active tunnels', true],
      ['10 concurrent connections', true],
      ['2 GB monthly bandwidth', true],
      ['3-day request retention', true],
      ['Custom domains', false],
      ['Team workspace', false],
    ],
  },
  {
    name: 'Link',
    description: 'For independent developers shipping previews.',
    usd: 7,
    ngn: 10_000,
    featured: false,
    features: [
      ['3 active tunnels', true],
      ['50 concurrent connections', true],
      ['25 GB monthly bandwidth', true],
      ['14-day request retention', true],
      ['1 custom domain', true],
      ['3 workspace members', true],
    ],
  },
  {
    name: 'Route',
    description: 'For teams testing APIs and shared environments.',
    usd: 15,
    ngn: 21_000,
    featured: true,
    features: [
      ['5 active tunnels', true],
      ['100 concurrent connections', true],
      ['100 GB monthly bandwidth', true],
      ['30-day request retention', true],
      ['5 custom domains', true],
      ['Priority support', true],
    ],
  },
  {
    name: 'Edge',
    description: 'For sustained traffic and larger organizations.',
    usd: 120,
    ngn: 170_000,
    featured: false,
    features: [
      ['20 active tunnels', true],
      ['500 concurrent connections', true],
      ['1 TB monthly bandwidth', true],
      ['90-day request retention', true],
      ['25 custom domains', true],
      ['Unlimited members', true],
    ],
  },
] as const;

export function PricingPage() {
  const [yearly, setYearly] = useState(false);

  return (
    <section className="pb-20 pt-28 sm:pt-32">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <div className="inline-flex rounded-full border border-white/10 bg-white/4 px-4 py-2 text-sm text-white/55">
            Predictable plans, visible limits
          </div>
          <h1 className="mt-6 text-4xl font-bold tracking-tighter sm:text-5xl lg:whitespace-nowrap">
            Pay for capacity, not surprises.
          </h1>
          <p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-white/50">
            Start at no cost, then add traffic, retention, domains, and team
            capacity as your workflow grows.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <div className="inline-flex rounded-full border border-white/10 bg-white/[0.035] p-1">
              <button
                type="button"
                onClick={() => setYearly(false)}
                className={`rounded-full px-4 py-2 text-sm transition-colors ${!yearly ? 'bg-white text-black' : 'text-white/45 hover:text-white'}`}
              >
                Monthly
              </button>
              <button
                type="button"
                onClick={() => setYearly(true)}
                className={`flex items-center gap-2 rounded-full px-4 py-2 text-sm transition-colors ${yearly ? 'bg-indigo-300 text-[#080914]' : 'text-white/45 hover:text-white'}`}
              >
                Yearly
                <span className="rounded-full bg-emerald-300 px-2 py-0.5 text-[10px] font-semibold text-emerald-950">
                  2 months free
                </span>
              </button>
            </div>
          </div>
        </div>

        <div className="mt-14 grid gap-5 md:grid-cols-2 xl:grid-cols-4">
          {plans.map((plan) => {
            const usdAmount = yearly ? plan.usd * 10 : plan.usd;
            const ngnAmount = yearly ? plan.ngn * 10 : plan.ngn;
            const period = yearly ? 'year' : 'month';
            const formattedUSD = new Intl.NumberFormat('en-US', {
              style: 'currency',
              currency: 'USD',
              maximumFractionDigits: 0,
            }).format(usdAmount);
            const formattedNGN = new Intl.NumberFormat('en-NG', {
              style: 'currency',
              currency: 'NGN',
              maximumFractionDigits: 0,
            }).format(ngnAmount);

            return (
              <article
                key={plan.name}
                className={`relative flex flex-col rounded-3xl border p-7 transition-transform hover:-translate-y-1 ${plan.featured ? 'border-indigo-300/55 bg-linear-to-br from-indigo-300/15 via-white/4 to-cyan-300/6 shadow-[0_0_55px_rgba(129,140,248,0.14)]' : 'border-white/10 bg-[#090909] hover:border-white/20'}`}
              >
                {plan.featured && (
                  <div className="absolute -top-3 left-1/2 flex -translate-x-1/2 items-center gap-1 rounded-full bg-indigo-300 px-3 py-1 text-[11px] font-bold uppercase tracking-wide text-[#080914]">
                    <Sparkles className="size-3" /> Best fit
                  </div>
                )}
                <h2 className="text-xl font-bold">{plan.name}</h2>
                <p className="mt-2 min-h-12 text-sm leading-6 text-white/45">
                  {plan.description}
                </p>
                <div className="mt-7 flex items-end gap-1">
                  <span className="text-4xl font-bold tracking-tight">
                    {formattedUSD}
                  </span>
                  {usdAmount > 0 && (
                    <span className="pb-1 text-sm text-white/35">
                      /{period}
                    </span>
                  )}
                </div>
                <p className="mt-2 text-sm text-white/40">
                  {formattedNGN} / {period} equivalent
                </p>
                <div className="my-7 h-px bg-white/10" />
                <ul className="flex-1 space-y-4">
                  {plan.features.map(([feature, included]) => (
                    <li
                      key={feature}
                      className={`flex items-center gap-3 text-sm ${included ? 'text-white/70' : 'text-white/25'}`}
                    >
                      <span
                        className={`flex size-5 shrink-0 items-center justify-center rounded-full ${included ? 'bg-indigo-300/12 text-indigo-200' : 'bg-white/4'}`}
                      >
                        {included ? (
                          <Check className="size-3" />
                        ) : (
                          <X className="size-3" />
                        )}
                      </span>
                      {feature}
                    </li>
                  ))}
                </ul>
                <Link
                  to="/signup"
                  className={`mt-8 rounded-full py-3 text-center text-sm font-bold transition-colors ${plan.featured ? 'bg-white text-black hover:bg-white/85' : 'bg-white/[0.07] text-white hover:bg-white/12'}`}
                >
                  {usdAmount === 0 ? 'Start free' : `Choose ${plan.name}`}
                </Link>
              </article>
            );
          })}
        </div>
      </MarketingContainer>
    </section>
  );
}
