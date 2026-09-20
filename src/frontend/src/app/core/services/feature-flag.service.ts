import { Injectable } from '@angular/core';

const FLAGS = {
  darkMode: false,
} as const;

export type FeatureFlag = keyof typeof FLAGS;

@Injectable({ providedIn: 'root' })
export class FeatureFlagService {
  isEnabled(flag: FeatureFlag): boolean {
    return FLAGS[flag];
  }
}
