import { TestBed, fakeAsync, tick } from '@angular/core/testing';
import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { slowRequestInterceptor, SLOW_REQUEST_THRESHOLD_MS } from './slow-request.interceptor';
import { SlowRequestService } from '../services/slow-request.service';

describe('slowRequestInterceptor', () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let slowRequestSpy: jasmine.SpyObj<SlowRequestService>;

  beforeEach(() => {
    slowRequestSpy = jasmine.createSpyObj('SlowRequestService', ['markSlow', 'markDone']);

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([slowRequestInterceptor])),
        provideHttpClientTesting(),
        { provide: SlowRequestService, useValue: slowRequestSpy },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('should not call markSlow when the request completes before the threshold', fakeAsync(() => {
    http.get('/api/fast').subscribe();

    tick(SLOW_REQUEST_THRESHOLD_MS - 1);
    httpMock.expectOne('/api/fast').flush({});
    tick();

    expect(slowRequestSpy.markSlow).not.toHaveBeenCalled();
    expect(slowRequestSpy.markDone).not.toHaveBeenCalled();
  }));

  it('should call markSlow after the threshold elapses', fakeAsync(() => {
    http.get('/api/slow').subscribe();

    tick(SLOW_REQUEST_THRESHOLD_MS);

    expect(slowRequestSpy.markSlow).toHaveBeenCalledTimes(1);

    httpMock.expectOne('/api/slow').flush({});
    tick();
  }));

  it('should call markDone after a slow request completes successfully', fakeAsync(() => {
    http.get('/api/slow').subscribe();

    tick(SLOW_REQUEST_THRESHOLD_MS);
    httpMock.expectOne('/api/slow').flush({});
    tick();

    expect(slowRequestSpy.markDone).toHaveBeenCalledTimes(1);
  }));

  it('should call markDone even when the request errors', fakeAsync(() => {
    http.get('/api/failing').subscribe({ error: () => {} });

    tick(SLOW_REQUEST_THRESHOLD_MS);
    httpMock.expectOne('/api/failing').flush('Server error', { status: 500, statusText: 'Internal Server Error' });
    tick();

    expect(slowRequestSpy.markDone).toHaveBeenCalledTimes(1);
  }));

  it('should not call markDone for a fast request that errors', fakeAsync(() => {
    http.get('/api/fast-fail').subscribe({ error: () => {} });

    tick(SLOW_REQUEST_THRESHOLD_MS - 1);
    httpMock.expectOne('/api/fast-fail').flush('Not found', { status: 404, statusText: 'Not Found' });
    tick();

    expect(slowRequestSpy.markSlow).not.toHaveBeenCalled();
    expect(slowRequestSpy.markDone).not.toHaveBeenCalled();
  }));

  it('should call markSlow independently for each concurrent slow request', fakeAsync(() => {
    http.get('/api/one').subscribe();
    http.get('/api/two').subscribe();

    tick(SLOW_REQUEST_THRESHOLD_MS);

    expect(slowRequestSpy.markSlow).toHaveBeenCalledTimes(2);

    httpMock.expectOne('/api/one').flush({});
    httpMock.expectOne('/api/two').flush({});
    tick();
  }));

  it('should call markDone for each concurrent slow request that completes', fakeAsync(() => {
    http.get('/api/one').subscribe();
    http.get('/api/two').subscribe();

    tick(SLOW_REQUEST_THRESHOLD_MS);

    httpMock.expectOne('/api/one').flush({});
    tick();
    expect(slowRequestSpy.markDone).toHaveBeenCalledTimes(1);

    httpMock.expectOne('/api/two').flush({});
    tick();
    expect(slowRequestSpy.markDone).toHaveBeenCalledTimes(2);
  }));

  it('should not call markSlow exactly at the threshold boundary', fakeAsync(() => {
    http.get('/api/boundary').subscribe();

    tick(SLOW_REQUEST_THRESHOLD_MS - 1);
    httpMock.expectOne('/api/boundary').flush({});
    tick(1);

    expect(slowRequestSpy.markSlow).not.toHaveBeenCalled();
  }));
});
