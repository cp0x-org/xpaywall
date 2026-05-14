import { ReactNode } from 'react';

// material-ui
import { styled, Theme, useTheme } from '@mui/material/styles';
import { MUIStyledCommonProps } from '@mui/system';
import Box from '@mui/material/Box';

// project import
import { ThemeDirection } from 'config';
import { withAlpha } from 'utils/colorUtils';

// third party
import { BrowserView, MobileView } from 'react-device-detect';
import SimpleBar, { Props } from 'simplebar-react';

// root style
const RootStyle = styled(BrowserView)({
  flexGrow: 1,
  height: '100%',
  overflow: 'hidden'
});

// scroll bar wrapper
const SimpleBarStyle = styled(SimpleBar)(({ theme }) => ({
  maxHeight: '100%',
  '& .simplebar-scrollbar': {
    '&:before': { backgroundColor: withAlpha(theme.vars.palette.grey[500], 0.48) },
    '&.simplebar-visible:before': { opacity: 1 }
  },
  '& .simplebar-track.simplebar-vertical': { width: 10 },
  '& .simplebar-track.simplebar-horizontal .simplebar-scrollbar': { height: 6 },
  '& .simplebar-mask': { zIndex: 'inherit' }
}));

// ==============================|| SIMPLE SCROLL BAR  ||============================== //

export default function SimpleBarScroll({ children, sx, ...other }: MUIStyledCommonProps<Theme> & Props) {
  const theme = useTheme();

  return (
    <>
      <RootStyle>
        <SimpleBarStyle
          clickOnTrack={false}
          sx={sx}
          data-simplebar-direction={theme.direction === ThemeDirection.RTL ? 'rtl' : 'ltr'}
          {...other}
        >
          {children as ReactNode}
        </SimpleBarStyle>
      </RootStyle>
      <MobileView>
        <Box sx={{ overflowX: 'auto', ...sx }} {...other}>
          {children as ReactNode}
        </Box>
      </MobileView>
    </>
  );
}
