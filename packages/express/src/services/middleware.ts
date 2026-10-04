import {
  createNodeHttpPayloadCaptureMiddleware,
  type HttpPayloadCaptureSetting,
} from '@outpipe/sdk';
import type { RequestHandler } from 'express';
import type { ExpressTunnel } from '../interfaces/options';

export function tunnelStatus(tunnel: ExpressTunnel): RequestHandler {
  return (_request, response) => {
    const state = tunnel.state();
    response.json({ ...state, error: state.error?.message });
  };
}

export function tunnelLifecycle(tunnel: ExpressTunnel): RequestHandler {
  return (_request, _response, next) => {
    void tunnel
      .start()
      .then(() => next())
      .catch(next);
  };
}

export function tunnelPayloadCapture(
  setting?: HttpPayloadCaptureSetting,
): RequestHandler {
  return createNodeHttpPayloadCaptureMiddleware(
    setting,
  ) as unknown as RequestHandler;
}
