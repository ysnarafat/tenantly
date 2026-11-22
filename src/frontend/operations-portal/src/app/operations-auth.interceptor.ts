import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';

const TOKEN_KEY = 'tenantly_ops_token';

export const operationsAuthInterceptor: HttpInterceptorFn = (request, next) => {
  if (request.url.includes('/auth/login')) {
    return next(request);
  }

  const token = sessionStorage.getItem(TOKEN_KEY);
  if (!token) {
    return next(request);
  }

  return next(request.clone({ setHeaders: { Authorization: `Bearer ${token}` } })).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 403 && error.error?.code === 'TOKEN_AUDIENCE_MISMATCH') {
        sessionStorage.removeItem(TOKEN_KEY);
      }
      return throwError(() => error);
    })
  );
};
