import '@angular/compiler';
import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting,
} from '@angular/common/http/testing';
import { getTestBed, TestBed } from '@angular/core/testing';
import {
  BrowserTestingModule,
  platformBrowserTesting,
} from '@angular/platform-browser/testing';
import { describe, expect, it } from 'vitest';
import { provideOutpipe } from '../providers';
import { OutpipeApiService } from '../services/outpipe-api.service';

getTestBed().initTestEnvironment(
  BrowserTestingModule,
  platformBrowserTesting(),
);

describe('OutpipeApiService', () => {
  it('builds an authenticated, encoded tunnel request through Angular DI', () => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideOutpipe({
          apiUrl: 'https://api.outpipe.dev/',
          apiKey: 'test-key',
        }),
      ],
    });

    const service = TestBed.inject(OutpipeApiService);
    const controller = TestBed.inject(HttpTestingController);
    let tunnels: unknown;

    service.listTunnels('org/team').subscribe((value) => {
      tunnels = value;
    });

    const request = controller.expectOne(
      'https://api.outpipe.dev/api/v1/organizations/org%2Fteam/tunnels',
    );
    expect(request.request.method).toBe('GET');
    expect(request.request.headers.get('Authorization')).toBe(
      'Bearer test-key',
    );
    request.flush([{ id: 'tunnel-1', status: 'active' }]);

    expect(tunnels).toEqual([{ id: 'tunnel-1', status: 'active' }]);
    controller.verify();
  });
});
