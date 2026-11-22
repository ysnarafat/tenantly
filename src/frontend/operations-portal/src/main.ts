import { provideZoneChangeDetection } from '@angular/core';
import { bootstrapApplication } from '@angular/platform-browser';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { OperationsApp } from './app/operations-app';
import { operationsAuthInterceptor } from './app/operations-auth.interceptor';

bootstrapApplication(OperationsApp, {
  providers: [
    provideZoneChangeDetection(),
    provideHttpClient(withInterceptors([operationsAuthInterceptor])),
  ],
}).catch((error: unknown) => console.error(error));
