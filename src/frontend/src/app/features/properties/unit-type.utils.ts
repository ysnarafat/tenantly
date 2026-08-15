import { BuildingType, UnitType } from '../../core/models';

export const UNIT_TYPES_BY_BUILDING_TYPE: Record<BuildingType, UnitType[]> = {
  Residential: ['Apartment', 'Parking', 'Storage'],
  Commercial: ['Shop', 'Office', 'Parking', 'Storage'],
  Mixed: ['Shop', 'Apartment', 'Office', 'Parking', 'Storage', 'Other'],
};

export function getUnitTypesForBuilding(buildingType: BuildingType): UnitType[] {
  return UNIT_TYPES_BY_BUILDING_TYPE[buildingType] || [];
}

export function getUnitTypeIcon(type: UnitType): string {
  switch (type) {
    case 'Shop':
      return 'store';
    case 'Apartment':
      return 'home';
    case 'Office':
      return 'business';
    case 'Parking':
      return 'local_parking';
    case 'Storage':
      return 'inventory_2';
    default:
      return 'meeting_room';
  }
}
