import { describe, expect, it, vi } from 'vitest';
import { createHoldToDeleteController } from './hold-to-delete-controller';

describe('createHoldToDeleteController', () => {
  it('increments progress up to 1 and confirms when completed', () => {
    let currentTime = 1000;
    let scheduledCallback: (() => void) | null = null;
    const triggerScheduled = () => {
      if (scheduledCallback) {
        scheduledCallback();
      }
    };
    const progressValues: number[] = [];
    const onConfirm = vi.fn();

    const controller = createHoldToDeleteController({
      durationMs: 1000,
      tickIntervalMs: 100,
      now: () => currentTime,
      schedule: (cb) => {
        scheduledCallback = cb;
        return 1;
      },
      cancelScheduled: () => {
        scheduledCallback = null;
      },
      onProgress: (p) => progressValues.push(p),
      onConfirm,
    });

    controller.start('pointer:1');
    expect(controller.isHolding()).toBe(true);
    expect(progressValues).toContain(0);

    // Simulate 500ms
    currentTime = 1500;
    triggerScheduled();
    expect(controller.getProgress()).toBe(0.5);
    expect(onConfirm).not.toHaveBeenCalled();

    // Simulate 1000ms
    currentTime = 2000;
    triggerScheduled();
    expect(controller.getProgress()).toBe(1);
    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(controller.isHolding()).toBe(false);
  });

  it('cancels early without calling onConfirm', () => {
    let currentTime = 1000;
    let scheduledCallback: (() => void) | null = null;
    const triggerScheduled = () => {
      if (scheduledCallback) {
        scheduledCallback();
      }
    };
    const onConfirm = vi.fn();

    const controller = createHoldToDeleteController({
      durationMs: 1000,
      tickIntervalMs: 100,
      now: () => currentTime,
      schedule: (cb) => {
        scheduledCallback = cb;
        return 1;
      },
      cancelScheduled: () => {
        scheduledCallback = null;
      },
      onConfirm,
    });

    controller.start('pointer:1');
    currentTime = 1500;
    triggerScheduled();
    expect(controller.getProgress()).toBe(0.5);

    controller.cancel('pointer:1');
    expect(controller.getProgress()).toBe(0);
    expect(controller.isHolding()).toBe(false);
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
