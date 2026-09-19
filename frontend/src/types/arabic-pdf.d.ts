declare module 'arabic-reshaper' {
  const ArabicReshaper: {
    convertArabic: (value: string) => string;
  };
  export default ArabicReshaper;
}

declare module 'bidi-js' {
  interface BidiResult {
    levels: Uint8Array;
    paragraphs: Array<{ start: number; end: number; level: number }>;
  }

  interface BidiProcessor {
    getEmbeddingLevels: (value: string, direction?: 'ltr' | 'rtl') => BidiResult;
    getReorderSegments: (value: string, levels: BidiResult, start?: number, end?: number) => Array<[number, number]>;
    getMirroredCharactersMap: (value: string, levels: BidiResult, start?: number, end?: number) => Map<number, string>;
  }

  const bidiFactory: () => BidiProcessor;
  export default bidiFactory;
}
