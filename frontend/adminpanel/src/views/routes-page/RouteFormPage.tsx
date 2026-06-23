import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import Typography from '@mui/material/Typography';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Grid from '@mui/material/Grid';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import LinkTwoToneIcon from '@mui/icons-material/LinkTwoTone';
import LanguageTwoToneIcon from '@mui/icons-material/LanguageTwoTone';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import useAuth from 'hooks/useAuth';
import axios from 'utils/axios';
import { canManage } from 'utils/ownership';

import { RouteRow } from './types';

interface Project {
  id: string;
  name: string;
  slug: string;
  base_url?: string | null;
  owner_user_id?: string | null;
  // Distinct protocols of the project's enabled payment methods (a project uses one).
  payment_methods?: string[];
}

interface ProxyUrlResp {
  proxy_url?: string;
}

function buildUrl(base: string, path: string): string {
  if (!base) return '';
  const cleanBase = base.replace(/\/$/, '');
  if (!path) return cleanBase;
  return cleanBase + (path.startsWith('/') ? path : `/${path}`);
}

const BAZAAR_TEMPLATE = {
  disabled: false,
  method: 'POST',
  body_type: 'json',
  input_example: { name: 'Alice', age: 30 },
  input_schema: {
    type: 'object',
    properties: {
      name: { type: 'string' },
      age: { type: 'integer' }
    },
    required: ['name']
  },
  output_example: { id: 'usr_1', name: 'Alice', age: 30 },
  output_schema: {
    type: 'object',
    properties: {
      id: { type: 'string' },
      name: { type: 'string' },
      age: { type: 'integer' }
    }
  }
};

const emptyValues = {
  project_id: '',
  name: '',
  path_pattern: '',
  free: false,
  price_usd: '',
  description: '',
  bazaar: '',
  submit: null
};

function inferJSONSchema(value: unknown): Record<string, unknown> {
  if (value === null) return { type: 'null' };
  if (Array.isArray(value)) {
    return {
      type: 'array',
      items: value.length > 0 ? inferJSONSchema(value[0]) : {}
    };
  }
  if (typeof value === 'object') {
    const props: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      props[k] = inferJSONSchema(v);
    }
    return { type: 'object', properties: props };
  }
  if (typeof value === 'number') {
    return { type: Number.isInteger(value) ? 'integer' : 'number' };
  }
  if (typeof value === 'boolean') return { type: 'boolean' };
  return { type: 'string' };
}

export default function RouteFormPage() {
  const { pathname, state } = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();
  const currentUserId = (user as { id?: string } | null | undefined)?.id;

  const isCreate = pathname.includes('/create');
  const isEdit = pathname.includes('/edit');
  const isView = pathname.includes('/view');

  const routeId: string | undefined = (state as any)?.id;

  const [projects, setProjects] = useState<Project[]>([]);
  const [proxyUrl, setProxyUrl] = useState('');
  const [loading, setLoading] = useState(isEdit || isView);
  const [loadError, setLoadError] = useState('');
  const [initialValues, setInitialValues] = useState(emptyValues);

  // Auto-generator state
  const [genUrl, setGenUrl] = useState('');
  const [genMethod, setGenMethod] = useState<'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'>('GET');
  const [genBody, setGenBody] = useState('');
  const [genBusy, setGenBusy] = useState(false);
  const [genError, setGenError] = useState('');

  let title = 'Route';
  if (isCreate) title = 'Create Route';
  if (isEdit) title = 'Update Route';
  if (isView) title = 'View Route';

  useEffect(() => {
    axios.get<Project[]>('/api/v1/projects/with-config').then((res) => setProjects(res.data ?? []));
    axios.get<ProxyUrlResp>('/api/v1/system/proxy-url').then((res) => setProxyUrl(res.data.proxy_url ?? ''));
  }, []);

  const ownedProjects = projects.filter((p) => canManage(currentUserId, p.owner_user_id));
  const projectOwnerMap: Record<string, string | null | undefined> = {};
  for (const p of projects) projectOwnerMap[p.id] = p.owner_user_id;
  const canEditCurrent = isCreate ? true : canManage(currentUserId, projectOwnerMap[initialValues.project_id]);
  const projectsForSelect = isCreate ? ownedProjects : projects;

  useEffect(() => {
    if ((isEdit || isView) && routeId) {
      axios
        .get<RouteRow>(`/api/v1/outbound-routes/${routeId}`)
        .then((res) => {
          const d = res.data;
          setInitialValues({
            project_id: d.project_id,
            name: d.name,
            path_pattern: d.path_pattern,
            free: d.free,
            price_usd: d.price_usd ?? '',
            description: d.description ?? '',
            bazaar: d.bazaar ? JSON.stringify(d.bazaar, null, 2) : '',
            submit: null
          });
        })
        .catch(() => setLoadError('Failed to load route data'))
        .finally(() => setLoading(false));
    }
  }, [isEdit, isView, routeId]);

  if (loading) {
    return (
      <MainCard title={title}>
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </MainCard>
    );
  }

  if (loadError) {
    return (
      <MainCard title={title}>
        <Typography color="error">{loadError}</Typography>
      </MainCard>
    );
  }

  return (
    <MainCard title={title}>
      <Formik
        enableReinitialize
        initialValues={initialValues}
        validationSchema={Yup.object().shape({
          project_id: Yup.string().required('Project is required'),
          name: Yup.string().required('Name is required'),
          path_pattern: Yup.string().required('Path pattern is required'),
          price_usd: Yup.string().when('free', {
            is: false,
            then: (s) => s.required('Price is required').matches(/^\d+(\.\d+)?$/, 'Must be a positive number, e.g. 0.10')
          }),
          bazaar: Yup.string().test('json', 'Bazaar must be valid JSON object', (val) => {
            if (!val || !val.trim()) return true;
            try {
              const parsed = JSON.parse(val);
              return typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed);
            } catch {
              return false;
            }
          })
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            let bazaarObj: Record<string, unknown> | null = null;
            if (values.bazaar && values.bazaar.trim()) {
              bazaarObj = JSON.parse(values.bazaar);
            }
            const payload: Record<string, unknown> = {
              project_id: values.project_id,
              name: values.name,
              path_pattern: values.path_pattern,
              free: values.free,
              price_amount: 0,
              price_usd: values.free ? '' : values.price_usd,
              description: values.description,
              bazaar: bazaarObj
            };
            if (isEdit && routeId) {
              await axios.put(`/api/v1/outbound-routes/${routeId}`, payload);
            } else {
              await axios.post('/api/v1/outbound-routes', payload);
            }
            navigate('/routes');
          } catch (err: any) {
            setStatus({ success: false });
            setErrors({ submit: err?.error || err?.message || 'Failed to save route' });
            setSubmitting(false);
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, setFieldValue, touched, values }) => {
          const handleGenerateBazaar = async () => {
            setGenError('');
            if (!genUrl.trim()) {
              setGenError('URL is required');
              return;
            }
            let parsedBody: unknown = undefined;
            if (genBody.trim() && genMethod !== 'GET' && genMethod !== 'DELETE') {
              try {
                parsedBody = JSON.parse(genBody);
              } catch {
                setGenError('Request body must be valid JSON');
                return;
              }
            }
            setGenBusy(true);
            try {
              const init: RequestInit = { method: genMethod };
              if (parsedBody !== undefined) {
                init.headers = { 'Content-Type': 'application/json' };
                init.body = JSON.stringify(parsedBody);
              }
              let res: Response;
              try {
                res = await fetch(genUrl, init);
              } catch (netErr: any) {
                // fetch throws TypeError("Failed to fetch") on CORS, DNS failure, mixed-content, etc.
                const msg = netErr?.message || String(netErr);
                const hint =
                  msg === 'Failed to fetch' || msg.toLowerCase().includes('network')
                    ? ' — likely CORS (server did not return Access-Control-Allow-Origin), DNS, or the target is unreachable. Check the browser devtools Network tab for the exact reason.'
                    : '';
                setGenError(`Request failed: ${msg}${hint}`);
                return;
              }
              const text = await res.text();
              if (!res.ok) {
                const snippet = text ? ` — ${text.slice(0, 200)}` : '';
                setGenError(`HTTP ${res.status} ${res.statusText}${snippet}`);
                return;
              }
              let outputExample: unknown = text;
              try {
                outputExample = JSON.parse(text);
              } catch {
                // keep raw text
              }
              const bazaar: Record<string, unknown> = {
                disabled: false,
                method: genMethod,
                body_type: parsedBody !== undefined ? 'json' : ''
              };
              if (parsedBody !== undefined) {
                bazaar.input_example = parsedBody;
                bazaar.input_schema = inferJSONSchema(parsedBody);
              }
              if (typeof outputExample === 'object' && outputExample !== null) {
                bazaar.output_example = outputExample;
                bazaar.output_schema = inferJSONSchema(outputExample);
              } else {
                bazaar.output_example = outputExample;
              }
              setFieldValue('bazaar', JSON.stringify(bazaar, null, 2));
            } finally {
              setGenBusy(false);
            }
          };

          const selectedProject = projects.find((p) => p.id === values.project_id);
          // Bazaar is an x402-only discovery extension — hide it for MPP projects.
          const isMppProject = (selectedProject?.payment_methods ?? []).includes('mpp');
          const fullProxyUrl = selectedProject
            ? buildUrl(
                proxyUrl,
                `/${selectedProject.slug}${values.path_pattern.startsWith('/') ? values.path_pattern : values.path_pattern ? `/${values.path_pattern}` : ''}`
              )
            : '';
          const fullTargetUrl = selectedProject ? buildUrl(selectedProject.base_url ?? '', values.path_pattern) : '';

          return (
            <form onSubmit={handleSubmit}>
              <Stack spacing={2} sx={{ maxWidth: 720 }}>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <Grid container spacing={1} sx={{ alignItems: 'center' }}>
                      <Grid>
                        <LinkTwoToneIcon color="primary" />
                      </Grid>
                      <Grid size={{ sm: 'grow' }}>
                        <Typography variant="h5" sx={{ wordBreak: 'break-all' }}>
                          {selectedProject ? fullProxyUrl || '—' : 'choose project'}
                        </Typography>
                        <Typography variant="subtitle2">PROXY URL</Typography>
                      </Grid>
                    </Grid>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <Grid container spacing={1} sx={{ alignItems: 'center' }}>
                      <Grid>
                        <LanguageTwoToneIcon color="secondary" />
                      </Grid>
                      <Grid size={{ sm: 'grow' }}>
                        <Typography variant="h5" sx={{ wordBreak: 'break-all' }}>
                          {selectedProject ? fullTargetUrl || '—' : 'choose project'}
                        </Typography>
                        <Typography variant="subtitle2">TARGET URL</Typography>
                      </Grid>
                    </Grid>
                  </Grid>
                </Grid>
                <Divider />

                <FormControl fullWidth error={Boolean(touched.project_id && errors.project_id)} disabled={isView}>
                  <InputLabel id="project-label">Project</InputLabel>
                  <Select
                    labelId="project-label"
                    name="project_id"
                    value={values.project_id}
                    label="Project"
                    onChange={(e) => setFieldValue('project_id', e.target.value)}
                  >
                    {projectsForSelect.map((p) => (
                      <MenuItem key={p.id} value={p.id}>
                        {p.name} ({p.slug})
                      </MenuItem>
                    ))}
                  </Select>
                  {touched.project_id && errors.project_id && <FormHelperText>{errors.project_id}</FormHelperText>}
                </FormControl>

                <TextField
                  fullWidth
                  label="Route Name"
                  name="name"
                  value={values.name}
                  onBlur={handleBlur}
                  onChange={handleChange}
                  error={Boolean(touched.name && errors.name)}
                  helperText={touched.name && errors.name}
                  disabled={isView}
                />

                <TextField
                  fullWidth
                  label="Path Pattern"
                  name="path_pattern"
                  value={values.path_pattern}
                  onBlur={handleBlur}
                  onChange={handleChange}
                  error={Boolean(touched.path_pattern && errors.path_pattern)}
                  helperText={(touched.path_pattern && errors.path_pattern) || 'e.g. /api/v1/* or /users/:id'}
                  disabled={isView}
                />

                <TextField
                  fullWidth
                  multiline
                  minRows={2}
                  label="Description"
                  name="description"
                  value={values.description}
                  onBlur={handleBlur}
                  onChange={handleChange}
                  disabled={isView}
                />

                <FormControlLabel
                  control={
                    <Checkbox
                      name="free"
                      checked={values.free}
                      onChange={(e) => setFieldValue('free', e.target.checked)}
                      disabled={isView}
                    />
                  }
                  label="Free (no payment required)"
                />

                {!values.free && (
                  <TextField
                    fullWidth
                    label="Price (USD)"
                    name="price_usd"
                    value={values.price_usd}
                    onBlur={handleBlur}
                    onChange={handleChange}
                    error={Boolean(touched.price_usd && errors.price_usd)}
                    helperText={(touched.price_usd && errors.price_usd) || 'e.g. 0.10'}
                    disabled={isView}
                  />
                )}

                {!isMppProject && (
                  <>
                    <Divider />

                    <Accordion defaultExpanded={Boolean(values.bazaar)}>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Stack>
                          <Typography variant="subtitle1">Bazaar Discovery Extension</Typography>
                          <Typography variant="caption" color="text.secondary">
                            Optional JSON describing the endpoint's method, body type, request/response schemas and examples. Leave empty
                            for auto-mode (minimal GET declaration).
                          </Typography>
                        </Stack>
                      </AccordionSummary>
                      <AccordionDetails>
                        <Stack spacing={2}>
                          {!isView && (
                            <Stack direction="row" spacing={1} flexWrap="wrap">
                              <Button
                                size="small"
                                variant="outlined"
                                onClick={() => setFieldValue('bazaar', JSON.stringify(BAZAAR_TEMPLATE, null, 2))}
                              >
                                Insert Template
                              </Button>
                              <Button
                                size="small"
                                variant="outlined"
                                color="warning"
                                onClick={() => setFieldValue('bazaar', '')}
                                disabled={!values.bazaar}
                              >
                                Clear
                              </Button>
                              <Button
                                size="small"
                                variant="outlined"
                                onClick={() => {
                                  try {
                                    const parsed = JSON.parse(values.bazaar);
                                    setFieldValue('bazaar', JSON.stringify(parsed, null, 2));
                                  } catch {
                                    // ignore — validation will catch it
                                  }
                                }}
                                disabled={!values.bazaar}
                              >
                                Format
                              </Button>
                            </Stack>
                          )}

                          <TextField
                            fullWidth
                            multiline
                            minRows={8}
                            maxRows={24}
                            label="Bazaar JSON"
                            name="bazaar"
                            value={values.bazaar}
                            onBlur={handleBlur}
                            onChange={handleChange}
                            error={Boolean(touched.bazaar && errors.bazaar)}
                            helperText={touched.bazaar && errors.bazaar}
                            disabled={isView}
                            slotProps={{
                              input: { style: { fontFamily: 'monospace', fontSize: 13 } }
                            }}
                          />

                          {!isView && (
                            <Accordion variant="outlined">
                              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                                <Stack>
                                  <Typography variant="subtitle2">Auto-generate from a sample request</Typography>
                                  <Typography variant="caption" color="text.secondary">
                                    Sends the request from your browser, then derives Bazaar JSON from the request/response. Target must
                                    allow CORS from this origin.
                                  </Typography>
                                </Stack>
                              </AccordionSummary>
                              <AccordionDetails>
                                <Stack spacing={2}>
                                  <Stack direction="row" spacing={1}>
                                    <FormControl sx={{ minWidth: 120 }} size="small">
                                      <InputLabel id="gen-method-label">Method</InputLabel>
                                      <Select
                                        labelId="gen-method-label"
                                        value={genMethod}
                                        label="Method"
                                        onChange={(e) => setGenMethod(e.target.value as typeof genMethod)}
                                      >
                                        {['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((m) => (
                                          <MenuItem key={m} value={m}>
                                            {m}
                                          </MenuItem>
                                        ))}
                                      </Select>
                                    </FormControl>
                                    <TextField
                                      fullWidth
                                      size="small"
                                      label="Sample URL"
                                      placeholder="https://upstream.example.com/users"
                                      value={genUrl}
                                      onChange={(e) => setGenUrl(e.target.value)}
                                    />
                                  </Stack>

                                  {genMethod !== 'GET' && genMethod !== 'DELETE' && (
                                    <TextField
                                      fullWidth
                                      multiline
                                      minRows={3}
                                      size="small"
                                      label="Request body (JSON, optional)"
                                      value={genBody}
                                      onChange={(e) => setGenBody(e.target.value)}
                                      slotProps={{
                                        input: { style: { fontFamily: 'monospace', fontSize: 13 } }
                                      }}
                                    />
                                  )}

                                  {genError && <FormHelperText error>{genError}</FormHelperText>}

                                  <Box>
                                    <Button
                                      variant="contained"
                                      size="small"
                                      onClick={handleGenerateBazaar}
                                      disabled={genBusy}
                                      startIcon={genBusy ? <CircularProgress size={16} color="inherit" /> : undefined}
                                    >
                                      {genBusy ? 'Fetching…' : 'Get Bazaar Description'}
                                    </Button>
                                  </Box>
                                </Stack>
                              </AccordionDetails>
                            </Accordion>
                          )}
                        </Stack>
                      </AccordionDetails>
                    </Accordion>
                  </>
                )}

                {errors.submit && (
                  <Box>
                    <FormHelperText error>{errors.submit}</FormHelperText>
                  </Box>
                )}

                <Stack direction="row" spacing={2} justifyContent="flex-end">
                  <Button variant="outlined" onClick={() => navigate('/routes')} disabled={isSubmitting}>
                    {isView ? 'Back' : 'Cancel'}
                  </Button>
                  {!isView && canEditCurrent && (
                    <Button type="submit" variant="contained" disabled={isSubmitting}>
                      {isEdit ? 'Save Changes' : 'Create Route'}
                    </Button>
                  )}
                </Stack>
              </Stack>
            </form>
          );
        }}
      </Formik>
    </MainCard>
  );
}
