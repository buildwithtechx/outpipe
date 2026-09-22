import { PluginCodeSection } from './plugin-code-section';
import { PluginCta } from './plugin-cta';
import type { PluginDefinition } from './plugin-data';
import { PluginHero } from './plugin-hero';
import { PluginStackSection } from './plugin-stack-section';

export function PluginDetailPage({ plugin }: { plugin: PluginDefinition }) {
  return (
    <>
      <PluginHero plugin={plugin} />
      <PluginCodeSection plugin={plugin} />
      <PluginStackSection plugin={plugin} />
      <PluginCta plugin={plugin} />
    </>
  );
}
