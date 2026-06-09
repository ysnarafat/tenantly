import { TestBed } from '@angular/core/testing';
import { MatSnackBar } from '@angular/material/snack-bar';
import { SlowRequestService } from './slow-request.service';

describe('SlowRequestService', () => {
  let service: SlowRequestService;
  let snackBarSpy: jasmine.SpyObj<MatSnackBar>;
  let snackBarRefSpy: jasmine.SpyObj<{ dismiss: () => void }>;

  beforeEach(() => {
    snackBarRefSpy = jasmine.createSpyObj('MatSnackBarRef', ['dismiss']);
    snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);
    snackBarSpy.open.and.returnValue(snackBarRefSpy as any);

    TestBed.configureTestingModule({
      providers: [SlowRequestService, { provide: MatSnackBar, useValue: snackBarSpy }],
    });

    service = TestBed.inject(SlowRequestService);
  });

  describe('markSlow()', () => {
    it('should open the snackbar on the first slow request', () => {
      service.markSlow();

      expect(snackBarSpy.open).toHaveBeenCalledOnceWith(
        '⏳ This is taking longer than expected…',
        undefined,
        { panelClass: 'slow-request-snackbar' }
      );
    });

    it('should not open additional snackbars for concurrent slow requests', () => {
      service.markSlow();
      service.markSlow();
      service.markSlow();

      expect(snackBarSpy.open).toHaveBeenCalledTimes(1);
    });
  });

  describe('markDone()', () => {
    it('should dismiss the snackbar when the last slow request completes', () => {
      service.markSlow();
      service.markDone();

      expect(snackBarRefSpy.dismiss).toHaveBeenCalledTimes(1);
    });

    it('should not dismiss while other slow requests are still in-flight', () => {
      service.markSlow();
      service.markSlow();

      service.markDone(); // count drops to 1, not zero

      expect(snackBarRefSpy.dismiss).not.toHaveBeenCalled();
    });

    it('should dismiss only after all concurrent slow requests complete', () => {
      service.markSlow();
      service.markSlow();

      service.markDone();
      expect(snackBarRefSpy.dismiss).not.toHaveBeenCalled();

      service.markDone();
      expect(snackBarRefSpy.dismiss).toHaveBeenCalledTimes(1);
    });

    it('should not throw when called without a prior markSlow', () => {
      expect(() => service.markDone()).not.toThrow();
      expect(snackBarRefSpy.dismiss).not.toHaveBeenCalled();
    });

    it('should allow a new snackbar after all requests complete and a new slow one starts', () => {
      service.markSlow();
      service.markDone(); // resets to 0

      service.markSlow(); // second wave

      expect(snackBarSpy.open).toHaveBeenCalledTimes(2);
    });

    it('should not dismiss more than once for a single slow request', () => {
      service.markSlow();
      service.markDone();
      service.markDone(); // extra call — counter is already 0

      expect(snackBarRefSpy.dismiss).toHaveBeenCalledTimes(1);
    });
  });
});
