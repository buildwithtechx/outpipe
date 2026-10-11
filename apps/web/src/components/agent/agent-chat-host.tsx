import { Dialog as DialogPrimitive } from 'radix-ui';
import { useAgentChatState } from './agent-chat-context';
import { AgentChatPanel } from './agent-chat-panel';

export function AgentChatHost() {
  const { state, dispatch } = useAgentChatState();
  const close = () => dispatch({ type: 'close' });

  return (
    <DialogPrimitive.Root
      open={state.panelOpen}
      onOpenChange={(open) => {
        if (!open) close();
      }}
    >
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <DialogPrimitive.Content
          aria-describedby={undefined}
          className="fixed inset-y-0 right-0 z-50 flex h-full w-full max-w-2xl flex-col border-l border-white/10 bg-zinc-950 p-0 text-white shadow-2xl transition ease-in-out data-[state=closed]:duration-200 data-[state=open]:duration-300 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right"
        >
          <DialogPrimitive.Title className="sr-only">
            Troubleshooting Agent
          </DialogPrimitive.Title>
          <AgentChatPanel onClose={close} />
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
