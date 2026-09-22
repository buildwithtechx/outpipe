import { describe, expect, it } from 'vitest';
import { OutpipeModule } from '../services/tunnel.module';
import { OutpipeService } from '../services/tunnel.service';

describe('OutpipeModule', () => {
  it('registers its options and lifecycle service', () => {
    const dynamicModule = OutpipeModule.forRoot({
      relayUrl: 'wss://relay.test',
      agentToken: 'token',
      localPort: 3000,
    });
    expect(dynamicModule.module).toBe(OutpipeModule);
    expect(dynamicModule.exports).toContain(OutpipeService);
    expect(dynamicModule.providers).toHaveLength(2);
  });
});
