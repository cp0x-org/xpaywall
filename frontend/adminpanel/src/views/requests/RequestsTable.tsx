import * as React from 'react';

// material-ui
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// project imports
import RequestsTableHeader from './RequestsTableHeader';

// types
import { ArrangementOrder, KeyedObject } from 'types';
import { RequestLog } from './types';

function descendingComparator(a: KeyedObject, b: KeyedObject, orderBy: string) {
  if (b[orderBy] < a[orderBy]) return -1;
  if (b[orderBy] > a[orderBy]) return 1;
  return 0;
}

function getComparator(order: ArrangementOrder, orderBy: string) {
  return order === 'desc'
    ? (a: KeyedObject, b: KeyedObject) => descendingComparator(a, b, orderBy)
    : (a: KeyedObject, b: KeyedObject) => -descendingComparator(a, b, orderBy);
}

function stableSort(array: RequestLog[], comparator: (a: RequestLog, b: RequestLog) => number) {
  const stabilized = array.map((el, index) => [el, index] as const);
  stabilized.sort((a, b) => {
    const order = comparator(a[0], b[0]);
    if (order !== 0) return order;
    return a[1] - b[1];
  });
  return stabilized.map((el) => el[0]);
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString();
}

function methodColor(method: string): 'default' | 'primary' | 'warning' | 'error' | 'success' | 'info' {
  switch (method.toUpperCase()) {
    case 'GET': return 'success';
    case 'POST': return 'primary';
    case 'PUT': return 'warning';
    case 'PATCH': return 'info';
    case 'DELETE': return 'error';
    default: return 'default';
  }
}

function statusColor(code?: number): 'success' | 'warning' | 'error' | 'default' {
  if (!code) return 'default';
  if (code < 300) return 'success';
  if (code < 500) return 'warning';
  return 'error';
}

export default function RequestsTable({ rows }: { rows: RequestLog[] }) {
  const [order, setOrder] = React.useState<ArrangementOrder>('desc');
  const [orderBy, setOrderBy] = React.useState<string>('created_at');
  const [page, setPage] = React.useState<number>(0);
  const [rowsPerPage, setRowsPerPage] = React.useState<number>(25);

  const handleRequestSort = (_event: React.SyntheticEvent<Element, Event>, property: string) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  const handleChangePage = (_event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement> | undefined) => {
    if (event?.target.value) {
      setRowsPerPage(parseInt(event.target.value, 10));
    }
    setPage(0);
  };

  const emptyRows = page > 0 ? Math.max(0, (1 + page) * rowsPerPage - rows.length) : 0;

  return (
    <>
      <TableContainer>
        <Table sx={{ minWidth: 900 }} aria-labelledby="requestsTable">
          <RequestsTableHeader order={order} orderBy={orderBy} onRequestSort={handleRequestSort} rowCount={rows.length} />
          <TableBody>
            {stableSort(rows, getComparator(order, orderBy))
              .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
              .map((row) => (
                <TableRow hover tabIndex={-1} key={row.id}>
                  <TableCell sx={{ whiteSpace: 'nowrap', color: 'text.secondary', fontSize: '0.8rem' }}>
                    {formatTime(row.created_at)}
                  </TableCell>

                  <TableCell>
                    <Chip label={row.method} size="small" color={methodColor(row.method)} sx={{ fontWeight: 700, fontSize: 11 }} />
                  </TableCell>

                  <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.82rem', maxWidth: 260 }}>
                    <Tooltip title={row.path}>
                      <Typography
                        sx={{
                          fontFamily: 'monospace',
                          fontSize: '0.82rem',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          maxWidth: 260
                        }}
                      >
                        {row.path}
                      </Typography>
                    </Tooltip>
                  </TableCell>

                  <TableCell align="right">
                    {row.final_status_code ? (
                      <Chip
                        label={row.final_status_code}
                        size="small"
                        color={statusColor(row.final_status_code)}
                        sx={{ fontWeight: 700, fontSize: 11 }}
                      />
                    ) : (
                      <Typography variant="caption" color="text.disabled">—</Typography>
                    )}
                  </TableCell>

                  <TableCell align="right">
                    {row.payment_completed ? (
                      <Chip label="Paid" size="small" color="success" sx={{ fontWeight: 600, fontSize: 11 }} />
                    ) : row.payment_required ? (
                      <Chip label="Required" size="small" color="warning" sx={{ fontWeight: 600, fontSize: 11 }} />
                    ) : (
                      <Typography variant="caption" color="text.disabled">—</Typography>
                    )}
                  </TableCell>

                  <TableCell align="right" sx={{ fontFamily: 'monospace', fontSize: '0.82rem' }}>
                    {row.amount_usd ? `$${parseFloat(row.amount_usd).toFixed(4)}` : <Typography variant="caption" color="text.disabled">—</Typography>}
                  </TableCell>

                  <TableCell align="right" sx={{ fontSize: '0.82rem' }}>
                    {row.upstream_response_time_ms != null
                      ? `${row.upstream_response_time_ms} ms`
                      : <Typography variant="caption" color="text.disabled">—</Typography>}
                  </TableCell>

                  <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.8rem', color: 'text.secondary' }}>
                    {row.client_ip ?? '—'}
                  </TableCell>
                </TableRow>
              ))}
            {emptyRows > 0 && (
              <TableRow sx={{ height: 53 * emptyRows }}>
                <TableCell colSpan={8} />
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <TablePagination
        rowsPerPageOptions={[10, 25, 50, 100]}
        component="div"
        count={rows.length}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </>
  );
}
