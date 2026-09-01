import {
  ArrowRight,
  Check,
  Handshake,
  Mail,
  MessageCircle,
  ShieldCheck,
} from 'lucide-react';
import { motion } from 'motion/react';
import { type FormEvent, useState } from 'react';
import { MarketingContainer } from '#/components/layout';

const channels = [
  {
    label: 'Product help',
    detail: 'Setup, accounts, and tunnel questions',
    email: 'hello@outpipe.dev',
    icon: Mail,
  },
  {
    label: 'Partnerships',
    detail: 'Integrations and business conversations',
    email: 'partners@outpipe.dev',
    icon: Handshake,
  },
  {
    label: 'Security',
    detail: 'Private vulnerability disclosure',
    email: 'security@outpipe.dev',
    icon: ShieldCheck,
  },
] as const;

export function ContactPage() {
  const [copied, setCopied] = useState<string>();
  const [submitted, setSubmitted] = useState(false);

  async function copyEmail(email: string) {
    await navigator.clipboard.writeText(email);
    setCopied(email);
    setTimeout(() => setCopied(undefined), 1800);
  }

  function openEmailDraft(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const name = String(data.get('name') ?? '').trim();
    const email = String(data.get('email') ?? '').trim();
    const topic = String(data.get('topic') ?? '').trim();
    const message = String(data.get('message') ?? '').trim();
    const subject = encodeURIComponent(topic || `Outpipe message from ${name}`);
    const body = encodeURIComponent(
      `${message}\n\nFrom: ${name}\nReply to: ${email}`,
    );
    setSubmitted(true);
    window.location.href = `mailto:hello@outpipe.dev?subject=${subject}&body=${body}`;
  }

  return (
    <section className="pb-20 pt-28 sm:pt-32">
      <MarketingContainer className="max-w-5xl">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          className="mx-auto max-w-2xl text-center"
        >
          <div className="mx-auto flex size-16 items-center justify-center rounded-2xl border border-indigo-300/20 bg-indigo-300/10">
            <MessageCircle className="size-8 text-indigo-200" />
          </div>
          <h1 className="mt-6 text-4xl font-bold tracking-[-0.05em] sm:text-6xl">
            Start a conversation.
          </h1>
          <p className="mx-auto mt-4 max-w-xl text-lg leading-8 text-white/50">
            Tell us where your tunnel workflow is stuck, what you are building,
            or where Outpipe should connect next.
          </p>
        </motion.div>

        <div className="mt-12 grid gap-4 md:grid-cols-3">
          {channels.map((channel, index) => (
            <motion.article
              key={channel.email}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.08 }}
              className="rounded-2xl border border-white/10 bg-[#090909] p-5"
            >
              <div className="flex items-start gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-white/[0.05]">
                  <channel.icon className="size-5 text-white/55" />
                </span>
                <div className="min-w-0">
                  <h2 className="font-semibold">{channel.label}</h2>
                  <p className="mt-1 text-xs leading-5 text-white/35">
                    {channel.detail}
                  </p>
                </div>
              </div>
              <div className="mt-5 flex items-center justify-between gap-2 border-t border-white/5 pt-4">
                <a
                  href={`mailto:${channel.email}`}
                  className="truncate text-xs font-medium text-indigo-200 hover:text-white"
                >
                  {channel.email}
                </a>
                <button
                  type="button"
                  onClick={() => copyEmail(channel.email)}
                  className="rounded-lg bg-white/[0.04] px-2.5 py-1.5 text-[11px] text-white/40 hover:bg-white/[0.08] hover:text-white"
                >
                  {copied === channel.email ? 'Copied' : 'Copy'}
                </button>
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
                <h2 className="text-lg font-semibold">Email draft opened</h2>
                <p className="mt-1 text-sm leading-6 text-white/45">
                  Review the message in your mail application, then send it when
                  you are ready.
                </p>
              </div>
            </div>
          ) : (
            <form onSubmit={openEmailDraft} className="grid gap-6">
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
              <label className="grid gap-2 text-sm text-white/55">
                Topic
                <input
                  required
                  name="topic"
                  placeholder="What would you like to discuss?"
                  className="rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                />
              </label>
              <label className="grid gap-2 text-sm text-white/55">
                Message
                <textarea
                  required
                  name="message"
                  rows={6}
                  placeholder="Share the useful context, without credentials or private tunnel data."
                  className="resize-none rounded-2xl border border-white/5 bg-black/45 px-4 py-3.5 text-white outline-hidden placeholder:text-white/20 focus:border-indigo-300/45 focus:ring-1 focus:ring-indigo-300/30"
                />
              </label>
              <button
                type="submit"
                className="inline-flex w-fit items-center gap-2 rounded-full bg-white px-6 py-3 text-sm font-bold text-black transition-transform hover:-translate-y-0.5"
              >
                Open email draft <ArrowRight className="size-4" />
              </button>
            </form>
          )}
        </motion.div>
      </MarketingContainer>
    </section>
  );
}
