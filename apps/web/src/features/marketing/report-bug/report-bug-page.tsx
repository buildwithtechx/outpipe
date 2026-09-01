import {
  ArrowRight,
  Bug,
  Check,
  FileCode2,
  ShieldCheck,
  Terminal,
} from 'lucide-react';
import { motion } from 'motion/react';
import { type FormEvent, useState } from 'react';
import { MarketingContainer } from '#/components/layout';

const categories = [
  'CLI or desktop client',
  'Tunnel connection',
  'Dashboard or billing',
  'SDK or framework integration',
  'Other',
];

const guidance = [
  {
    label: 'Useful context',
    detail: 'Include versions, commands, and the protocol involved.',
    icon: Terminal,
  },
  {
    label: 'Reproducible steps',
    detail: 'A short path from setup to failure helps us reproduce it quickly.',
    icon: FileCode2,
  },
  {
    label: 'Keep it safe',
    detail: 'Remove credentials, private URLs, and customer data from logs.',
    icon: ShieldCheck,
  },
] as const;

export function ReportBugPage() {
  const [submitted, setSubmitted] = useState(false);

  function openReport(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const name = String(data.get('name') ?? '').trim();
    const email = String(data.get('email') ?? '').trim();
    const category = String(data.get('category') ?? '').trim();
    const summary = String(data.get('summary') ?? '').trim();
    const reproduction = String(data.get('reproduction') ?? '').trim();
    const expected = String(data.get('expected') ?? '').trim();
    const actual = String(data.get('actual') ?? '').trim();
    const subject = encodeURIComponent(`[Bug] ${summary}`);
    const body = encodeURIComponent(
      [
        `From: ${name}`,
        `Reply-to: ${email}`,
        `Category: ${category}`,
        '',
        'Steps to reproduce:',
        reproduction,
        '',
        'Expected:',
        expected,
        '',
        'Actual:',
        actual,
      ].join('\n'),
    );

    setSubmitted(true);
    window.location.href = `mailto:bugs@outpipe.dev?subject=${subject}&body=${body}`;
  }

  return (
    <section className="pb-20 pt-28 sm:pt-32">
      <MarketingContainer className="max-w-5xl">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          className="mx-auto max-w-2xl text-center"
        >
          <div className="mx-auto flex size-16 items-center justify-center rounded-2xl border border-rose-300/25 bg-rose-300/10">
            <Bug className="size-8 text-rose-200" />
          </div>
          <h1 className="mt-6 text-4xl font-bold tracking-tighter sm:text-6xl">
            Help us fix it.
          </h1>
          <p className="mx-auto mt-4 max-w-xl text-lg leading-8 text-white/50">
            Found something that is not working as expected? Send the details
            and we’ll trace it from the first request to the last hop.
          </p>
        </motion.div>

        <div className="mt-12 grid gap-4 md:grid-cols-3">
          {guidance.map((item, index) => (
            <motion.article
              key={item.label}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.08 }}
              className="rounded-2xl border border-white/10 bg-[#090909] p-5"
            >
              <div className="flex items-start gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-white/5">
                  <item.icon className="size-5 text-white/55" />
                </span>
                <div>
                  <h2 className="font-semibold">{item.label}</h2>
                  <p className="mt-1 text-xs leading-5 text-white/35">
                    {item.detail}
                  </p>
                </div>
              </div>
            </motion.article>
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="mt-8 rounded-3xl border border-white/10 bg-[#090909] p-6 sm:p-10"
        >
          {submitted ? (
            <div className="flex items-start gap-4 rounded-2xl border border-emerald-300/20 bg-emerald-300/[0.07] p-6">
              <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-emerald-300/10">
                <Check className="size-5 text-emerald-300" />
              </span>
              <div>
                <h2 className="text-lg font-semibold">Report prepared</h2>
                <p className="mt-1 text-sm leading-6 text-white/45">
                  Your email client should be open with the report ready to
                  review and send.
                </p>
              </div>
            </div>
          ) : (
            <form onSubmit={openReport} className="grid gap-6">
              <div className="grid gap-5 md:grid-cols-2">
                <label className="grid gap-2 text-sm text-white/55">
                  Your name
                  <input
                    required
                    name="name"
                    placeholder="Ada Lovelace"
                    className="rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                  />
                </label>
                <label className="grid gap-2 text-sm text-white/55">
                  Email address
                  <input
                    required
                    type="email"
                    name="email"
                    placeholder="ada@example.com"
                    className="rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                  />
                </label>
              </div>
              <div className="grid gap-5 md:grid-cols-2">
                <label className="grid gap-2 text-sm text-white/55">
                  Area affected
                  <select
                    required
                    name="category"
                    defaultValue=""
                    className="rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                  >
                    <option value="" disabled>
                      Select an area
                    </option>
                    {categories.map((category) => (
                      <option key={category}>{category}</option>
                    ))}
                  </select>
                </label>
                <label className="grid gap-2 text-sm text-white/55">
                  Short summary
                  <input
                    required
                    name="summary"
                    placeholder="The tunnel disconnects after the first request"
                    className="rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                  />
                </label>
              </div>
              <label className="grid gap-2 text-sm text-white/55">
                Steps to reproduce
                <textarea
                  required
                  name="reproduction"
                  rows={5}
                  placeholder="1. Run ...  2. Open ...  3. Request ..."
                  className="resize-none rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                />
              </label>
              <div className="grid gap-5 md:grid-cols-2">
                <label className="grid gap-2 text-sm text-white/55">
                  Expected result
                  <textarea
                    required
                    name="expected"
                    rows={4}
                    placeholder="The request reaches my local service"
                    className="resize-none rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                  />
                </label>
                <label className="grid gap-2 text-sm text-white/55">
                  What happened instead
                  <textarea
                    required
                    name="actual"
                    rows={4}
                    placeholder="The connection closes with ..."
                    className="resize-none rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                  />
                </label>
              </div>
              <button
                type="submit"
                className="inline-flex w-fit items-center gap-2 rounded-full bg-white px-6 py-3 text-sm font-bold text-black transition-transform hover:-translate-y-0.5"
              >
                Prepare bug report <ArrowRight className="size-4" />
              </button>
            </form>
          )}
        </motion.div>
      </MarketingContainer>
    </section>
  );
}
