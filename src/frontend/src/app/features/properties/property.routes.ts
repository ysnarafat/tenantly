import { Routes } from '@angular/router';
import { provideState } from '@ngrx/store';
import { provideEffects } from '@ngrx/effects';
import { propertyReducer } from './store/property.reducer';
import { PropertyEffects } from './store/property.effects';
import { buildingReducer } from './store/building.reducer';
import { BuildingEffects } from './store/building.effects';
import { unitReducer } from './store/unit.reducer';
import { UnitEffects } from './store/unit.effects';
import { PropertyListComponent } from './property-list/property-list.component';

export const PROPERTY_ROUTES: Routes = [
    {
        path: '',
        component: PropertyListComponent,
        providers: [
            provideState('properties', propertyReducer),
            provideState('buildings', buildingReducer),
            provideState('units', unitReducer),
            provideEffects(PropertyEffects, BuildingEffects, UnitEffects)
        ]
    },
    // Child routes for details, buildings, etc. will be added here
];
