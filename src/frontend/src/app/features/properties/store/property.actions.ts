import { createActionGroup, emptyProps, props } from '@ngrx/store';
import { Property, CreatePropertyRequest, UpdatePropertyRequest } from '../../../core/models';

export const PropertyActions = createActionGroup({
    source: 'Property',
    events: {
        'Load Properties': props<{ active?: boolean }>(),
        'Load Properties Success': props<{ properties: Property[] }>(),
        'Load Properties Failure': props<{ error: string }>(),

        'Load Property': props<{ id: number }>(),
        'Load Property Success': props<{ property: Property }>(),
        'Load Property Failure': props<{ error: string }>(),

        'Create Property': props<{ property: CreatePropertyRequest }>(),
        'Create Property Success': props<{ property: Property }>(),
        'Create Property Failure': props<{ error: string }>(),

        'Update Property': props<{ id: number; property: UpdatePropertyRequest }>(),
        'Update Property Success': props<{ property: Property }>(),
        'Update Property Failure': props<{ error: string }>(),

        'Delete Property': props<{ id: number }>(),
        'Delete Property Success': props<{ id: number }>(),
        'Delete Property Failure': props<{ error: string }>(),

        'Select Property': props<{ id: number }>(),
        'Clear Selected Property': emptyProps(),
    }
});
