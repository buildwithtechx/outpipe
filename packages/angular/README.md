# @outpipe/angular

Angular integration for the Outpipe API. It provides an injectable API service
and standalone application providers while leaving tunnel protocol types in
`@outpipe/sdk`.

## Installation

```sh
npm install @outpipe/angular
```

## Configuration

Register the provider in a standalone Angular application:

```ts
import { provideHttpClient } from '@angular/common/http';
import { bootstrapApplication } from '@angular/platform-browser';
import { provideOutpipe } from '@outpipe/angular';
import { AppComponent } from './app.component';

bootstrapApplication(AppComponent, {
  providers: [
    provideHttpClient(),
    provideOutpipe({
      apiUrl: 'https://api.outpipe.dev',
      apiKey: 'agent-or-user-key',
    }),
  ],
});
```

## Use the service

```ts
import { Component, inject } from '@angular/core';
import { OutpipeApiService } from '@outpipe/angular';

@Component({
  selector: 'app-tunnels',
  template: '{{ (tunnels | async)?.length ?? 0 }} tunnels',
})
export class TunnelsComponent {
  private readonly outpipe = inject(OutpipeApiService);
  readonly tunnels = this.outpipe.listTunnels('organization-id');
}
```

The adapter currently covers tunnel creation, listing, inspection, and
revocation. Use `@outpipe/sdk` directly when an application needs relay
connection or protocol operations.
