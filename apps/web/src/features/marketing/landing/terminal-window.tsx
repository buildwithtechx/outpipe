import { AnimatePresence, motion, useReducedMotion } from 'motion/react';
import { useEffect, useState } from 'react';

const terminalRequestSequence = [
  ['GET', '/api/health', '200', '12ms'],
  ['POST', '/webhooks', '201', '45ms'],
  ['GET', '/oauth/callback', '302', '8ms'],
  ['GET', '/favicon.ico', '304', '2ms'],
  ['GET', '/api/projects', '200', '24ms'],
  ['POST', '/api/uploads', '201', '230ms'],
  ['DELETE', '/api/sessions/123', '204', '35ms'],
  ['PATCH', '/api/profile', '200', '67ms'],
  ['POST', '/api/checkout', '500', '120ms'],
] as const;

function statusColor(status: string) {
  const code = Number(status);

  if (code >= 500) return 'text-red-300';
  if (code >= 400) return 'text-orange-300';
  if (code >= 300) return 'text-amber-300';
  return 'text-emerald-300';
}

export function TerminalWindow() {
  const reducedMotion = useReducedMotion();
  const [visibleRequests, setVisibleRequests] = useState(() =>
    terminalRequestSequence.slice(0, 8).map((request, id) => ({ request, id })),
  );

  useEffect(() => {
    if (reducedMotion) return;
    let nextRequest = 8;
    const timer = setInterval(() => {
      setVisibleRequests((current) => [
        ...current.slice(1),
        {
          request:
            terminalRequestSequence[
              nextRequest % terminalRequestSequence.length
            ],
          id: nextRequest++,
        },
      ]);
    }, 1400);

    return () => clearInterval(timer);
  }, [reducedMotion]);

  return (
    <div className="mt-10 w-full min-w-0 max-w-5xl overflow-hidden rounded-[1.25rem] border border-white/15 bg-[#090a0c] text-left font-mono text-sm shadow-2xl shadow-indigo-950/50 sm:mt-14">
      <div className="relative flex items-center gap-2 border-b border-white/10 bg-white/6 px-4 py-4 sm:gap-3 sm:px-6">
        <span className="size-2.5 shrink-0 rounded-full bg-red-400 sm:size-3" />
        <span className="size-2.5 shrink-0 rounded-full bg-amber-300 sm:size-3" />
        <span className="size-2.5 shrink-0 rounded-full bg-emerald-400 sm:size-3" />
        <span className="ml-2 min-w-0 truncate text-xs text-white/35 sm:absolute sm:inset-x-0 sm:ml-0 sm:text-center sm:text-sm">
          user@outpipe-cli
        </span>
      </div>
      <div className="grid min-w-0 grid-cols-1 gap-2 px-4 py-6 text-[11px] leading-6 [overflow-wrap:anywhere] sm:px-8 sm:py-8 sm:text-sm">
        <p className="text-white/85">
          <span className="text-emerald-300">➜</span>{' '}
          <span className="text-cyan-300">~</span> outpipe open --port 3000
        </p>
        <p className="text-cyan-300">Connecting to Outpipe...</p>
        <p className="text-emerald-300">Linked to local port 3000</p>
        <p className="text-fuchsia-300">
          Tunnel ready: https://quiet-moon.outpipe.app
        </p>
        <p className="text-amber-300">
          Keep this process running to keep the tunnel active.
        </p>
        <div className="mt-3 grid gap-2 text-white/45">
          <AnimatePresence initial={false} mode="popLayout">
            {visibleRequests.map(({ request, id }) => (
              <motion.div
                key={id}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                className="grid grid-cols-[3rem_minmax(0,1fr)_2rem_2.75rem] items-center gap-2 sm:grid-cols-[4.5rem_minmax(0,1fr)_3.5rem_3.5rem] sm:gap-3"
              >
                <span>{request[0]}</span>
                <span className="truncate text-white/65">{request[1]}</span>
                <span className={`text-right ${statusColor(request[2])}`}>
                  {request[2]}
                </span>
                <span className="text-right text-white/30">{request[3]}</span>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
