import { useState } from 'react';
import { CopyCommand } from '#/components/ui/copy-command';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { docsUrl } from '#/lib/docs';

export function CliInstall() {
  const [platform, setPlatform] = useState('unix');
  return (
    <Tabs value={platform} onValueChange={setPlatform}>
      <TabsList aria-label="Installation platform">
        <TabsTrigger value="unix">macOS / Linux</TabsTrigger>
        <TabsTrigger value="windows">Windows</TabsTrigger>
      </TabsList>
      <TabsContent value="unix">
        <CopyCommand
          text="curl -fsSL https://cli.outpipe.dev | bash"
          label="Copy macOS and Linux install command"
        />
        <p className="text-xs text-muted-foreground">
          Add the installed CLI directory to PATH, then run{' '}
          <code>outpipe version</code>.
        </p>
      </TabsContent>
      <TabsContent value="windows">
        <p className="text-sm text-muted-foreground">
          Download the Windows archive, extract <code>outpipe.exe</code> and add
          its directory to PATH.
        </p>
        <a
          href="https://github.com/buildwithtechx/outpipe/releases"
          target="_blank"
          rel="noopener noreferrer"
          className="mt-3 inline-block text-sm text-indigo-200 underline"
        >
          Download Windows CLI ↗
        </a>
      </TabsContent>
      <a
        href={docsUrl('installation')}
        className="mt-3 inline-block text-xs text-muted-foreground underline"
      >
        Installation guide
      </a>
    </Tabs>
  );
}
