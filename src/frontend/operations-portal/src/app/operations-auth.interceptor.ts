import { HttpInterceptorFn } from '@angular/common/http';

const TOKEN_KEY = 'tenantly_ops_token';

export const operationsAuthInterceptor: HttpInterceptorFn = (request, next) => {
  if (request.url.includes('/auth/login')) {
    return next(request);
  }

  const token = sessionStorage.getItem(TOKEN_KEY);
  if (!token) {
    return next(request);
  }

  return next(request.clone({ setHeaders: { Authorization: `Bearer ${token}` } }));
};
