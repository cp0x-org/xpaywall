import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import CardContent from '@mui/material/CardContent';
import InputAdornment from '@mui/material/InputAdornment';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import EntityTable from './EntityTable';

// assets
import AddIcon from '@mui/icons-material/AddTwoTone';
import SearchIcon from '@mui/icons-material/Search';

// types
import { EntityPageConfig, EntityRow } from './types';

export default function EntityListPage({ config }: { config: EntityPageConfig }) {
  const [search, setSearch] = useState<string>('');

  const rows = useMemo(() => {
    if (!search.trim()) return config.rows;

    const query = search.toLowerCase();
    return config.rows.filter((row: EntityRow) => {
      return (
        row.name.toLowerCase().includes(query) ||
        row.owner.toLowerCase().includes(query) ||
        row.status.toLowerCase().includes(query) ||
        row.id.toString().includes(query)
      );
    });
  }, [search, config.rows]);

  return (
    <MainCard content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <TextField
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={`Search ${config.singularName}`}
            size="small"
            sx={{ width: { xs: 1, sm: 'auto' } }}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" />
                  </InputAdornment>
                )
              }
            }}
          />

          <Button component={Link} to={`${config.basePath}/create`} variant="contained" startIcon={<AddIcon fontSize="small" />}>
            Create {config.singularName}
          </Button>
        </Stack>
      </CardContent>
      <EntityTable rows={rows} basePath={config.basePath} />
    </MainCard>
  );
}
