import { Link } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { ConfirmAction } from '#/components/feedback/confirm-action';
import { UnsavedChangesGuard } from '#/components/feedback/unsaved-changes-guard';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { Textarea } from '#/components/ui/textarea';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { ApiError } from '#/lib/api-client';
import {
  useMonitors,
  useStatusPageConfig,
  useUptimeMutations,
} from './hooks/use-uptime';
import {
  publicStatusUrl,
  validStatusDomain,
  validStatusSlug,
} from './lib/status-page-settings';
import { StatusPagePublishing } from './status-page-publishing';

export function StatusPageEditor({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const orgId = organizationQuery.organization?.id;
  const query = useStatusPageConfig(orgId);
  const monitors = useMonitors(orgId);
  const { saveStatusPage } = useUptimeMutations(orgId ?? '');
  const [slug, setSlug] = useState('');
  const [title, setTitle] = useState('Service status');
  const [description, setDescription] = useState('');
  const [domain, setDomain] = useState('');
  const [published, setPublished] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [notice, setNotice] = useState('');
  const missing = query.error instanceof ApiError && query.error.status === 404;
  useEffect(() => {
    if (dirty || !orgId) return;
    setSlug(query.data?.slug ?? orgSlug);
    setTitle(query.data?.title ?? 'Service status');
    setDescription(query.data?.description ?? '');
    setDomain(query.data?.customDomain ?? '');
    setPublished(query.data?.isPublic ?? false);
  }, [query.data, orgSlug, orgId, dirty]);
  const statusBase =
    import.meta.env.VITE_OUTPIPE_STATUS_URL ||
    (import.meta.env.DEV
      ? 'http://localhost:4322'
      : 'https://status.outpipe.dev');
  const url = publicStatusUrl(slug, statusBase, domain);
  const valid =
    validStatusSlug(slug) &&
    Boolean(title.trim()) &&
    title.length <= 255 &&
    description.length <= 1024 &&
    validStatusDomain(domain);
  const edit = () => {
    setDirty(true);
    setNotice('');
    saveStatusPage.reset();
  };
  async function save() {
    if (!valid || !orgId) return;
    try {
      await saveStatusPage.mutateAsync({
        slug,
        title: title.trim(),
        description: description.trim(),
        customDomain: domain,
        is_public: published,
      });
      setDirty(false);
      setNotice(
        published
          ? 'Status page published.'
          : 'Draft saved. Your page is private.',
      );
    } catch (error) {
      setNotice('Could not save this page. Check the settings and try again.');
      throw error;
    }
  }
  if (organizationQuery.isLoading || query.isLoading)
    return <p role="status">Loading status page…</p>;
  if (!orgId && !organizationQuery.isError)
    return (
      <p role="alert">Workspace not found. Choose an available workspace.</p>
    );
  if (organizationQuery.isError || (query.isError && !missing))
    return (
      <div role="alert">
        <p>Could not load the status page.</p>
        <Button
          onClick={() =>
            void (organizationQuery.isError
              ? organizationQuery.refetch()
              : query.refetch())
          }
          variant="outline"
        >
          Try again
        </Button>
      </div>
    );
  return (
    <div className="space-y-6">
      <UnsavedChangesGuard dirty={dirty} />
      <header className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-6">
        <div>
          <Link
            to="/$orgSlug/uptime"
            params={{ orgSlug }}
            className="text-sm text-muted-foreground"
          >
            ← Uptime
          </Link>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight">
            Status page
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Prepare, preview and publish a clear view of your service health.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-muted-foreground">
            {dirty
              ? 'Unsaved changes'
              : query.data?.isPublic
                ? 'Published'
                : 'Draft'}
          </span>
          {query.data?.isPublic && (slug !== query.data.slug || !published) ? (
            <ConfirmAction
              title="Change your public status page?"
              description="Unpublishing hides your page. Changing its slug makes the previous public link unavailable."
              label="Save changes"
              pending={saveStatusPage.isPending}
              disabled={!valid}
              onConfirm={save}
            />
          ) : (
            <Button
              disabled={
                !valid ||
                saveStatusPage.isPending ||
                (!dirty && Boolean(query.data))
              }
              onClick={() => void save().catch(() => {})}
            >
              {saveStatusPage.isPending ? 'Saving…' : 'Save changes'}
            </Button>
          )}
        </div>
      </header>
      {notice && (
        <p
          role={saveStatusPage.isError ? 'alert' : 'status'}
          className={`rounded-xl border border-border p-4 text-sm ${saveStatusPage.isError ? 'text-rose-300' : 'text-emerald-300'}`}
        >
          {notice}
        </p>
      )}
      <fieldset disabled={saveStatusPage.isPending} className="contents">
        <Tabs defaultValue="content" className="space-y-6">
          <TabsList
            aria-label="Status page settings"
            className="h-auto flex-wrap"
          >
            <TabsTrigger value="content">Content</TabsTrigger>
            <TabsTrigger value="services">Services</TabsTrigger>
            <TabsTrigger value="preview">Preview</TabsTrigger>
            <TabsTrigger value="publishing">Publishing & domain</TabsTrigger>
          </TabsList>
          <TabsContent
            value="content"
            className="max-w-2xl space-y-5 rounded-2xl border border-border bg-card p-6"
          >
            <div className="space-y-2">
              <Label htmlFor="status-title">Page title</Label>
              <Input
                id="status-title"
                maxLength={255}
                value={title}
                onChange={(event) => {
                  edit();
                  setTitle(event.target.value);
                }}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="status-description">Description</Label>
              <Textarea
                id="status-description"
                maxLength={1024}
                rows={4}
                value={description}
                onChange={(event) => {
                  edit();
                  setDescription(event.target.value);
                }}
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Your public page uses the same Outpipe branding, typography and
              accessible controls as the dashboard.
            </p>
          </TabsContent>
          <TabsContent value="services" className="space-y-4">
            <h2 className="font-semibold">Services shown on your page</h2>
            <p className="text-sm text-muted-foreground">
              All workspace monitors and incident updates appear on the
              published page.
            </p>
            {monitors.isLoading ? (
              <p role="status">Loading services…</p>
            ) : monitors.isError ? (
              <p role="alert">Could not load services.</p>
            ) : monitors.data?.length ? (
              <ul className="divide-y divide-border rounded-xl border border-border">
                {monitors.data.map((monitor) => (
                  <li
                    key={monitor.id}
                    className="flex flex-wrap items-center justify-between gap-3 p-4"
                  >
                    <span>{monitor.name}</span>
                    <span className="text-sm capitalize text-muted-foreground">
                      {monitor.status}
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted-foreground">
                Add a monitor in Uptime to show service health.
              </p>
            )}
            <Link
              to="/$orgSlug/uptime"
              params={{ orgSlug }}
              className="inline-block text-sm text-indigo-200 underline"
            >
              Manage monitors and incidents
            </Link>
          </TabsContent>
          <TabsContent value="preview">
            <section
              aria-label="Status page content preview"
              className="space-y-4 rounded-2xl border border-border bg-card p-6 sm:p-10"
            >
              <p className="text-xs font-medium uppercase tracking-wider text-indigo-200">
                Outpipe status · Content preview
              </p>
              <h2 className="text-3xl font-semibold">
                {title || 'Service status'}
              </h2>
              <p className="text-muted-foreground">
                {description || 'Current availability and incident updates.'}
              </p>
              <p className="rounded-xl border border-dashed border-border p-4 text-sm text-muted-foreground">
                {monitors.data?.length ?? 0} monitored services. This preview
                reflects your unsaved content; visit the published page for live
                health and incident updates.
              </p>
            </section>
          </TabsContent>
          <StatusPagePublishing
            slug={slug}
            domain={domain}
            published={published}
            url={url}
            statusBase={statusBase}
            saved={query.data}
            onChange={(changes) => {
              edit();
              if (changes.slug !== undefined) setSlug(changes.slug);
              if (changes.domain !== undefined) setDomain(changes.domain);
              if (changes.published !== undefined)
                setPublished(changes.published);
            }}
          />
        </Tabs>
      </fieldset>
      {!valid && (
        <p role="alert" className="text-sm text-amber-300">
          Use a valid URL slug, non-empty title and hostname before saving.
        </p>
      )}
    </div>
  );
}
