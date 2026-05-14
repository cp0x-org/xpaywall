export interface EntityRow {
  id: number;
  name: string;
  owner: string;
  status: 'Active' | 'Planned' | 'Completed';
  startDate: string;
  dueDate: string;
  budget: number;
}

export interface EntityPageConfig {
  basePath: string;
  singularName: string;
  pluralName: string;
  rows: EntityRow[];
}
