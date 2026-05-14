import EntityListPage from 'views/entity-pages/EntityListPage';
import { STAT_ROWS } from './data';

export default function StatsPage() {
  return (
    <EntityListPage
      config={{
        basePath: '/stats',
        singularName: 'Stat',
        pluralName: 'Stats',
        rows: STAT_ROWS
      }}
    />
  );
}
