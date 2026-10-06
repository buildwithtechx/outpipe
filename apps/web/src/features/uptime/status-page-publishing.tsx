import { CopyCommand } from '#/components/ui/copy-command';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { TabsContent } from '#/components/ui/tabs';
import type { UptimeStatusPage } from '#/interfaces/uptime';
import {
  publicStatusUrl,
  validStatusDomain,
  validStatusSlug,
} from './lib/status-page-settings';

type PublishingDraft = { slug: string; domain: string; published: boolean };

export function StatusPagePublishing({
  slug,
  domain,
  published,
  url,
  statusBase,
  saved,
  onChange,
}: PublishingDraft & {
  url: string | null;
  statusBase: string;
  saved?: UptimeStatusPage;
  onChange: (changes: Partial<PublishingDraft>) => void;
}) {
  const savedUrl = saved
    ? publicStatusUrl(saved.slug, statusBase, saved.customDomain)
    : null;
  return (
    <TabsContent
      value="publishing"
      className="max-w-2xl space-y-6 rounded-2xl border border-border bg-card p-6"
    >
      <div className="space-y-2">
        <Label htmlFor="status-slug">Public URL slug</Label>
        <Input
          id="status-slug"
          value={slug}
          aria-invalid={!validStatusSlug(slug)}
          onChange={(event) => {
            onChange({ slug: event.target.value.toLowerCase() });
          }}
        />
        <p className="text-xs text-muted-foreground">
          Use lowercase letters, numbers and hyphens. Changing a published slug
          changes its public link.
        </p>
      </div>
      {url ? (
        <CopyCommand text={url} label="Copy status page URL" />
      ) : (
        <p role="alert" className="text-sm text-amber-300">
          A public URL is unavailable. Check the slug, domain and status site
          URL configuration.
        </p>
      )}
      <div className="space-y-2">
        <Label htmlFor="status-domain">Custom domain (optional)</Label>
        <Input
          id="status-domain"
          placeholder="status.example.com"
          value={domain}
          aria-invalid={!validStatusDomain(domain)}
          onChange={(event) => {
            onChange({ domain: event.target.value.trim().toLowerCase() });
          }}
        />
        <p className="text-xs text-muted-foreground">
          A hostname only. DNS and HTTPS must be provisioned before this
          hostname can serve your page. Saving it does not verify domain
          ownership or configure DNS.
        </p>
      </div>
      <label className="flex min-h-11 items-center gap-3 rounded-xl border border-border p-4">
        <input
          type="checkbox"
          checked={published}
          onChange={(event) => {
            onChange({ published: event.target.checked });
          }}
        />
        <span className="text-sm">
          Publish this page for anyone with the link
        </span>
      </label>
      {saved?.isPublic && savedUrl && (
        <a
          href={savedUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-block text-sm text-indigo-200 underline"
        >
          View published page ↗
        </a>
      )}
    </TabsContent>
  );
}
