import browserCollections from 'fumadocs-mdx:collections/browser';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { createServerFn } from '@tanstack/react-start';
import { useFumadocsLoader } from 'fumadocs-core/source/client';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import defaultMdxComponents from 'fumadocs-ui/mdx';
import {
  DocsBody,
  DocsDescription,
  DocsPage,
  DocsTitle,
} from 'fumadocs-ui/page';
import type { ComponentType } from 'react';
import { createSeo } from '#/lib/seo';
import { source } from '#/lib/source';

export const Route = createFileRoute('/docs/$')({
  loader: async ({ params }) => {
    const slugs = params._splat ? params._splat.split('/') : [];
    const data = await serverLoader({ data: slugs });
    await clientLoader.preload(data.path);
    return data;
  },
  head: ({ loaderData }) =>
    createSeo({
      title: loaderData?.title
        ? `${loaderData.title} | Outpipe Docs`
        : 'Outpipe Documentation',
      description: loaderData?.description,
      path: loaderData?.path ?? '/docs',
    }),
  notFoundComponent: () => (
    <div className="flex min-h-[50vh] flex-col items-center justify-center p-8 text-center">
      <h1 className="text-2xl font-bold">404 - Document Not Found</h1>
      <p className="mt-2 text-muted-foreground">
        The requested documentation page could not be found.
      </p>
    </div>
  ),
  component: Page,
});

const serverLoader = createServerFn({ method: 'GET' })
  .validator((slugs: string[]) => slugs)
  .handler(async ({ data: slugs }) => {
    const page = source.getPage(slugs);
    if (!page) throw notFound();

    return {
      path: page.path,
      title: page.data.title,
      description: page.data.description,
      pageTree: await source.serializePageTree(source.getPageTree()),
    };
  });

const clientLoader = browserCollections.docs.createClientLoader({
  component({ toc, frontmatter, default: MDX }) {
    return (
      <DocsPage toc={toc}>
        <DocsTitle>{frontmatter.title}</DocsTitle>
        <DocsDescription>{frontmatter.description}</DocsDescription>
        <DocsBody>
          <MDX components={defaultMdxComponents} />
        </DocsBody>
      </DocsPage>
    );
  },
});

function Page() {
  const data = Route.useLoaderData();
  const { pageTree } = useFumadocsLoader(data);
  const Content = clientLoader.getComponent(
    data.path,
  ) as unknown as ComponentType;

  return (
    <DocsLayout tree={pageTree} nav={{ title: 'Outpipe' }}>
      <Content />
    </DocsLayout>
  );
}
