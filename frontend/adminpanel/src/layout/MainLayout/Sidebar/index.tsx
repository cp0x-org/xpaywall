import { memo, useMemo } from 'react';

// material-ui
import { Theme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Drawer from '@mui/material/Drawer';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';

// project imports
import MenuCard from './MenuCard';
import MenuList from '../MenuList';
import MiniDrawerStyled from './MiniDrawerStyled';

import { MenuOrientation } from 'config';
import useConfig from 'hooks/useConfig';
import { drawerWidth } from 'store/constant';
import SimpleBar from 'ui-component/third-party/SimpleBar';

import { handlerDrawerOpen, useGetMenuMaster } from 'api/menu';

// ==============================|| SIDEBAR DRAWER ||============================== //

function Sidebar() {
  const downMD = useMediaQuery((theme: Theme) => theme.breakpoints.down('md'));

  const { menuMaster } = useGetMenuMaster();
  const drawerOpen = menuMaster.isDashboardDrawerOpened;

  const {
    state: { menuOrientation, miniDrawer }
  } = useConfig();

  const drawer = useMemo(() => {
    const isVerticalOpen = menuOrientation === MenuOrientation.VERTICAL && drawerOpen;
    const drawerContent = <MenuCard />;

    let drawerSX = { paddingLeft: '0px', paddingRight: '0px' };
    if (drawerOpen) drawerSX = { paddingLeft: '16px', paddingRight: '16px' };

    const versionBadge = isVerticalOpen ? (
      <Box sx={{ px: 2, pb: 2, pt: 1 }}>
        <Chip
          label="BETA"
          size="small"
          sx={{
            height: 18,
            fontSize: '0.6rem',
            fontWeight: 700,
            letterSpacing: '0.05em',
            bgcolor: 'warning.main',
            color: 'warning.contrastText',
            '& .MuiChip-label': { px: 0.75 }
          }}
        />
        {/*<Box*/}
        {/*  sx={{*/}
        {/*    display: 'flex',*/}
        {/*    alignItems: 'center',*/}
        {/*    gap: 1,*/}
        {/*    px: 1.5,*/}
        {/*    py: 0.75,*/}
        {/*    borderRadius: 2,*/}
        {/*    bgcolor: 'primary.light',*/}
        {/*    border: '1px solid',*/}
        {/*    borderColor: 'primary.200'*/}
        {/*  }}*/}
        {/*>*/}
        {/*  /!*<Typography variant="caption" sx={{ color: 'text.secondary', fontSize: '0.7rem' }}>*!/*/}
        {/*  /!*  v0.1.0*!/*/}
        {/*  /!*</Typography>*!/*/}
        {/*  <Chip*/}
        {/*    label="BETA"*/}
        {/*    size="small"*/}
        {/*    sx={{*/}
        {/*      height: 18,*/}
        {/*      fontSize: '0.6rem',*/}
        {/*      fontWeight: 700,*/}
        {/*      letterSpacing: '0.05em',*/}
        {/*      bgcolor: 'warning.main',*/}
        {/*      color: 'warning.contrastText',*/}
        {/*      '& .MuiChip-label': { px: 0.75 }*/}
        {/*    }}*/}
        {/*  />*/}
        {/*</Box>*/}
      </Box>
    ) : null;

    return (
      <>
        {downMD ? (
          <Box sx={{ ...drawerSX, display: 'flex', flexDirection: 'column', height: '100%' }}>
            <Box sx={{ flexGrow: 1 }}>
              <MenuList />
              {isVerticalOpen && drawerContent}
            </Box>
            {versionBadge}
          </Box>
        ) : (
          <SimpleBar sx={{ height: 'calc(100vh - 56px)', ...drawerSX }}>
            <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: 'calc(100vh - 88px)' }}>
              <Box sx={{ flexGrow: 1 }}>
                <MenuList />
                {isVerticalOpen && drawerContent}
              </Box>
              {versionBadge}
            </Box>
          </SimpleBar>
        )}
      </>
    );
  }, [downMD, drawerOpen, menuOrientation]);

  return (
    <Box component="nav" sx={{ flexShrink: { md: 0 }, width: { xs: 'auto', md: drawerWidth } }} aria-label="mailbox folders">
      {downMD || (miniDrawer && drawerOpen) ? (
        <Drawer
          variant={downMD ? 'temporary' : 'persistent'}
          anchor="left"
          open={drawerOpen}
          onClose={() => handlerDrawerOpen(!drawerOpen)}
          slotProps={{
            paper: {
              sx: {
                mt: downMD ? 0 : 7,
                zIndex: 1099,
                width: drawerWidth,
                bgcolor: 'background.default',
                color: 'text.primary',
                borderRight: 'none'
              }
            }
          }}
          ModalProps={{ keepMounted: true }}
          color="inherit"
        >
          {drawer}
        </Drawer>
      ) : (
        <MiniDrawerStyled variant="permanent" open={drawerOpen}>
          {drawer}
        </MiniDrawerStyled>
      )}
    </Box>
  );
}

export default memo(Sidebar);
