// ==============================|| DATA GRID STYLE - HOOKS ||============================== //

export default function useDataGrid() {
  const dataGridStyles = {
    '& .MuiDataGrid-root .MuiDataGrid-columnSeparator': { color: 'grey.300' },
    '.MuiDataGrid-root .MuiDataGrid-container--top [role=row],': { backgroundColor: 'background.paper' }
  };

  return dataGridStyles;
}
