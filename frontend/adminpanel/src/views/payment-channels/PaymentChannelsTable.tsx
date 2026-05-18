import * as React from 'react';
import { Link } from 'react-router-dom';

// material-ui
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import PaymentChannelsTableHeader from './PaymentChannelsTableHeader';

// assets
import DeleteTwoToneIcon from '@mui/icons-material/DeleteTwoTone';
import EditTwoToneIcon from '@mui/icons-material/EditTwoTone';
import VisibilityTwoToneIcon from '@mui/icons-material/VisibilityTwoTone';

// types
import { ArrangementOrder, KeyedObject } from 'types';
import { PaymentChannel } from './types';

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

function stableSort(array: PaymentChannel[], comparator: (a: PaymentChannel, b: PaymentChannel) => number) {
  const stabilized = array.map((el: PaymentChannel, index: number) => [el, index] as const);
  stabilized.sort((a, b) => {
    const order = comparator(a[0], b[0]);
    if (order !== 0) return order;
    return a[1] - b[1];
  });
  return stabilized.map((el) => el[0]);
}

export default function PaymentChannelsTable({ rows }: { rows: PaymentChannel[] }) {
  const [order, setOrder] = React.useState<ArrangementOrder>('asc');
  const [orderBy, setOrderBy] = React.useState<string>('protocol');
  const [page, setPage] = React.useState<number>(0);
  const [rowsPerPage, setRowsPerPage] = React.useState<number>(10);

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
    <MainCard content={false}>
      <TableContainer>
        <Table sx={{ minWidth: 760 }} aria-labelledby="tableTitle">
          <PaymentChannelsTableHeader
            order={order}
            orderBy={orderBy}
            onRequestSort={handleRequestSort}
            rowCount={rows.length}
          />
          <TableBody>
            {stableSort(rows, getComparator(order, orderBy))
              .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
              .map((row) => (
                <TableRow hover tabIndex={-1} key={row.id}>
                  <TableCell>
                    <Typography variant="subtitle1">{row.protocol}</Typography>
                  </TableCell>
                  <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>{row.method}</TableCell>
                  <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>{row.scheme}</TableCell>
                  <TableCell>
                    <Chip label={row.enabled ? 'Enabled' : 'Disabled'} size="small" color={row.enabled ? 'success' : 'default'} />
                  </TableCell>
                  <TableCell>{new Date(row.created_at).toLocaleDateString()}</TableCell>
                  <TableCell align="center" sx={{ pr: 3 }}>
                    <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                      <Tooltip title="View">
                        <IconButton
                          color="primary"
                          component={Link}
                          to="/payment-channels/view"
                          state={{ id: row.id }}
                          size="small"
                          aria-label="View"
                        >
                          <VisibilityTwoToneIcon sx={{ fontSize: '1.3rem' }} />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit">
                        <IconButton
                          color="secondary"
                          component={Link}
                          to="/payment-channels/edit"
                          state={{ id: row.id }}
                          size="small"
                          aria-label="Edit"
                        >
                          <EditTwoToneIcon sx={{ fontSize: '1.3rem' }} />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton color="error" size="small" aria-label="Delete">
                          <DeleteTwoToneIcon sx={{ fontSize: '1.3rem' }} />
                        </IconButton>
                      </Tooltip>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            {emptyRows > 0 && (
              <TableRow sx={{ height: 53 * emptyRows }}>
                <TableCell colSpan={6} />
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <TablePagination
        rowsPerPageOptions={[5, 10, 25]}
        component="div"
        count={rows.length}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </MainCard>
  );
}
