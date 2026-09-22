import { InjectionToken } from '@angular/core';
import type { OutpipeAngularConfig } from './interfaces';

export const OUTPIPE_API_CONFIG = new InjectionToken<OutpipeAngularConfig>(
  'OUTPIPE_API_CONFIG',
);
