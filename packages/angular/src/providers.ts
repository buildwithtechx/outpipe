import {
  type EnvironmentProviders,
  makeEnvironmentProviders,
} from '@angular/core';
import type { OutpipeAngularConfig } from './interfaces';
import { OUTPIPE_API_CONFIG } from './tokens';

export function provideOutpipe(
  config: OutpipeAngularConfig,
): EnvironmentProviders {
  return makeEnvironmentProviders([
    { provide: OUTPIPE_API_CONFIG, useValue: config },
  ]);
}
