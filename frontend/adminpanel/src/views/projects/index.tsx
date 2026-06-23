import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

// material-ui
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';
import InputAdornment from '@mui/material/InputAdornment';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import ProjectsTable from './ProjectsTable';
import axiosServices from 'utils/axios';

// assets
import AddIcon from '@mui/icons-material/AddTwoTone';
import SearchIcon from '@mui/icons-material/Search';

// types
import { Project } from './types';

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>('');

  useEffect(() => {
    axiosServices
      .get<Project[]>('/api/v1/projects/with-config')
      .then((res) => setProjects(res.data ?? []))
      .catch((err) => setError(err?.error || err?.message || 'Failed to load projects'))
      .finally(() => setLoading(false));
  }, []);

  const rows = useMemo(() => {
    if (!search.trim()) return projects;
    const query = search.toLowerCase();
    return projects.filter((p) => p.name.toLowerCase().includes(query) || p.slug.toLowerCase().includes(query) || p.id.includes(query));
  }, [search, projects]);

  return (
    <MainCard content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <TextField
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search Project"
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

          <Button component={Link} to="/projects/create" variant="contained" startIcon={<AddIcon fontSize="small" />}>
            Create Project
          </Button>
        </Stack>
      </CardContent>

      {error && (
        <CardContent>
          <Alert severity="error">{error}</Alert>
        </CardContent>
      )}

      {loading ? (
        <CardContent sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress />
        </CardContent>
      ) : (
        <ProjectsTable
          rows={rows}
          onDelete={async (id) => {
            await axiosServices.delete(`/api/v1/projects/${id}`);
            setProjects((prev) => prev.filter((p) => p.id !== id));
          }}
        />
      )}
    </MainCard>
  );
}
