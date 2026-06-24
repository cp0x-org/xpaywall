import * as React from 'react';

// material-ui
import { visuallyHidden } from '@mui/utils';
import Box from '@mui/material/Box';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';

// types
import { ArrangementOrder, HeadCell } from 'types';

const headCells: HeadCell[] = [
  { id: 'name', numeric: false, label: 'Name' },
  { id: 'url', numeric: false, label: 'URL' },
  { id: 'enabled', numeric: false, label: 'Enabled' },
  { id: 'is_global', numeric: false, label: 'Global' },
  { id: 'created_at', numeric: false, label: 'Created At' }
];

interface FacilitatorsTableHeaderProps {
  order: ArrangementOrder;
  orderBy: string;
  rowCount: number;
  onRequestSort: (event: React.SyntheticEvent<Element, Event>, property: string) => void;
}

export default function FacilitatorsTableHeader({ order, orderBy, onRequestSort }: FacilitatorsTableHeaderProps) {
  const createSortHandler = (property: string) => (event: React.SyntheticEvent<Element, Event>) => {
    onRequestSort(event, property);
  };

  return (
    <TableHead>
      <TableRow>
        {headCells.map((headCell) => {
          const isActive = orderBy === headCell.id;
          return (
            <TableCell
              key={headCell.id}
              align={headCell.align}
              padding={headCell.disablePadding ? 'none' : 'normal'}
              sortDirection={isActive ? order : undefined}
            >
              <TableSortLabel active={isActive} direction={isActive ? order : 'asc'} onClick={createSortHandler(headCell.id)}>
                {headCell.label}
                {isActive && (
                  <Box component="span" sx={visuallyHidden}>
                    {order === 'desc' ? 'sorted descending' : 'sorted ascending'}
                  </Box>
                )}
              </TableSortLabel>
            </TableCell>
          );
        })}
        <TableCell sortDirection={false} align="center" sx={{ pr: 3 }}>
          Action
        </TableCell>
      </TableRow>
    </TableHead>
  );
}
