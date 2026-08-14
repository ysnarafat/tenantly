// Used by both the "production" and "container-dev" build configurations
// (angular.json) — anything built inside Docker sits behind nginx, which
// already proxies /api/ on the same origin, so both need this relative
// apiUrl regardless of optimization level. Only local `ng serve` (no nginx
// in front) needs the absolute URL in environment.ts.
export const environment = {
  production: true,
  apiUrl: '/api/v1',
};
