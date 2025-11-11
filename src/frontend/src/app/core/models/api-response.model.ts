import { Property } from './property.model';
import { Building } from './building.model';
import { Unit } from './unit.model';

export interface Pagination {
    current_page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
}

export interface PaginatedResponse<T> {
    data: T[];
    pagination: Pagination;
}

// Specific response for properties since the key is 'properties' not 'data'
export interface PropertyListResponse {
    properties: Property[];
    pagination: Pagination;
}

export interface BuildingListResponse {
    buildings: Building[];
    pagination: Pagination;
}

export interface UnitListResponse {
    units: Unit[];
    pagination: Pagination;
}
