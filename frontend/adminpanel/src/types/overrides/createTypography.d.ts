import '@mui/material/styles';

declare module '@mui/material/styles' {
  export interface FontStyle
    extends Required<{
      textTransform: TextTransform;
      fontSize: string | number; // added string
    }> {}
  export interface FontStyleOptions extends Partial<FontStyle> {
    fontSize?: string | number; // added string
  }
  export type Variant = 'commonAvatar' | 'smallAvatar' | 'mediumAvatar' | 'largeAvatar';

  export interface TypographyVariantsOptions extends Partial<Record<Variant, TypographyStyleOptions> & FontStyleOptions> {
    commonAvatar?: TypographyStyleOptions;
    smallAvatar?: TypographyStyleOptions;
    mediumAvatar?: TypographyStyleOptions;
    largeAvatar?: TypographyStyleOptions;
  }

  export interface TypographyVariants extends Record<Variant, TypographyStyle>, FontStyle, TypographyUtils {
    commonAvatar: TypographyStyle;
    smallAvatar: TypographyStyle;
    mediumAvatar: TypographyStyle;
    largeAvatar: TypographyStyle;
  }
}
