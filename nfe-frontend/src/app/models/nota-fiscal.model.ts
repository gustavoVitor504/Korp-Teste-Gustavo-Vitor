export enum StatusNotaFiscal {
  Aberta = 'aberta',
  Fechada = 'fechada',
  Cancelada = 'cancelada',
}

export interface ItemNotaFiscal {
  produtoId: number;
  descricao: string;
  quantidade: number;
}

export interface NotaFiscal {
  id: number;
  numero: number;
  numeroFormatado: string;
  status: StatusNotaFiscal;
  itens: ItemNotaFiscal[];
}

export interface ItemPayload {
  produtoId: number;
  quantidade: number;
}

// A nota sempre nasce como Aberta no backend — não dá pra escolher
// status na criação, então o payload só leva os itens.
export interface CriarNotaFiscalPayload {
  itens: ItemPayload[];
}

// Formatos de erro estruturado que o backend pode devolver.
export interface CorpoErroNegocio {
  erro: 'saldo_insuficiente' | 'estoque_indisponivel' | 'nota_nao_aberta';
  mensagem?: string;
  produto?: string;
  disponivel?: number;
  solicitado?: number;
}
