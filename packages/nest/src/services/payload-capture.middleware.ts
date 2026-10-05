import type { IncomingMessage, ServerResponse } from 'node:http';
import { Injectable, type NestMiddleware } from '@nestjs/common';
import { createNodeHttpPayloadCaptureMiddleware } from '@outpipe/sdk';

@Injectable()
export class OutpipePayloadCaptureMiddleware implements NestMiddleware {
  private readonly handler = createNodeHttpPayloadCaptureMiddleware();

  use(
    req: IncomingMessage,
    res: ServerResponse,
    next: (error?: unknown) => void,
  ): void {
    this.handler(req, res, next);
  }
}
