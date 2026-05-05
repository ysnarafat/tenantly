import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { timer } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { SlowRequestService } from '../services/slow-request.service';

export const SLOW_REQUEST_THRESHOLD_MS = 3000;

export const slowRequestInterceptor: HttpInterceptorFn = (req, next) => {
  const slowRequestService = inject(SlowRequestService);
  let markedSlow = false;

  const timerSub = timer(SLOW_REQUEST_THRESHOLD_MS).subscribe(() => {
    markedSlow = true;
    slowRequestService.markSlow();
  });

  return next(req).pipe(
    finalize(() => {
      timerSub.unsubscribe();
      if (markedSlow) {
        slowRequestService.markDone();
      }
    })
  );
};
