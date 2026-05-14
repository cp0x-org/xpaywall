// material-ui
import { Theme } from '@mui/material/styles';
import { ChipProps } from '@mui/material/Chip';

// project imports
import { withAlpha } from 'utils/colorUtils';

// ===============================||  OVERRIDES - CHIP  ||=============================== //

export default function Chip(theme: Theme) {
  return {
    MuiChip: {
      defaultProps: {
        color: 'primary' as ChipProps['color'],
        variant: 'light' as ChipProps['variant']
      },
      styleOverrides: {
        root: {
          variants: [
            {
              props: { variant: 'light' }, // Variant for light Chip
              style: ({ ownerState, theme }: { ownerState: ChipProps; theme: Theme }) => {
                // Make sure color exists and is a key of palette
                const colorKey = ownerState.color as keyof typeof theme.vars.palette;
                const paletteColor = theme.vars.palette[colorKey];

                if (!paletteColor) return {};

                return {
                  color: paletteColor.main,
                  backgroundColor: paletteColor.light,

                  ...(ownerState.color === 'error' && {
                    backgroundColor: withAlpha(paletteColor.light, 0.25)
                  }),
                  ...(ownerState.color === 'success' && {
                    backgroundColor: withAlpha(paletteColor.light, 0.5)
                  }),
                  ...((ownerState.color === 'warning' || ownerState.color === 'success') && {
                    color: paletteColor.dark
                  }),

                  // Dark mode styles
                  ...theme.applyStyles('dark', {
                    backgroundColor: withAlpha(paletteColor.dark, 0.15),
                    ...(ownerState.color === 'error' && {
                      backgroundColor: withAlpha(paletteColor.dark, 0.15)
                    }),
                    ...(ownerState.color === 'success' && {
                      backgroundColor: withAlpha(paletteColor.dark, 0.15)
                    }),
                    ...((ownerState.color === 'warning' || ownerState.color === 'success') && {
                      color: paletteColor.main
                    })
                  }),

                  '&.MuiChip-clickable': {
                    '&:hover': {
                      color: paletteColor.light,
                      backgroundColor: paletteColor.dark,

                      ...theme.applyStyles('dark', {
                        color: paletteColor.dark,
                        backgroundColor: paletteColor.light
                      })
                    }
                  }
                };
              }
            },
            {
              props: { variant: 'outlined', color: 'warning' },
              style: {
                borderColor: theme.vars.palette.warning.dark,
                color: theme.vars.palette.warning.dark
              }
            },
            {
              props: { variant: 'outlined', color: 'success' },
              style: {
                borderColor: theme.vars.palette.success.dark,
                color: theme.vars.palette.success.dark
              }
            }
          ],
          '&.MuiChip-deletable .MuiChip-deleteIcon': {
            color: 'inherit'
          }
        }
      }
    }
  };
}
