import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { ProdutoService } from '../../services/produto.service';
import {
  CorpoErroNegocio,
  CriarNotaFiscalPayload,
  ItemPayload,
  NotaFiscal,
  StatusNotaFiscal,
} from '../../models/nota-fiscal.model';
import { Produto } from '../../models/produto.model';

@Component({
  selector: 'app-nota-fiscal-form',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './nota-fiscal-form.component.html',
  styleUrl: './nota-fiscal-form.component.css',
})
export class NotaFiscalFormComponent implements OnInit {
  readonly StatusNotaFiscal = StatusNotaFiscal;

  produtosDisponiveis: Produto[] = [];
  notas: NotaFiscal[] = [];
  
  quantidades: Record<number, number> = {};

  carregando = false;
  salvando = false;
  imprimindoId: number | null = null;
  mensagem: { tipo: 'ok' | 'erro'; texto: string } | null = null;

  private notaFiscalService = inject(NotaFiscalService);
  private produtoService = inject(ProdutoService);

  ngOnInit(): void {
    this.carregarDados();
  }

  carregarDados(): void {
    this.carregando = true;
    this.produtoService.listar().subscribe({
      next: (produtos) => {
        this.produtosDisponiveis = produtos;
        this.carregando = false;
      },
      error: () => {
        this.carregando = false;
        this.mensagem = { tipo: 'erro', texto: 'Não foi possível carregar os produtos.' };
      },
    });

    this.notaFiscalService.listar().subscribe({
      next: (notas) => (this.notas = notas),
      error: () => {
        this.mensagem = { tipo: 'erro', texto: 'Não foi possível carregar as notas fiscais.' };
      },
    });
  }

  atualizarQuantidade(produtoId: number | undefined, valor: string): void {
    if (produtoId === undefined) return;

    const quantidade = Math.max(0, Math.floor(Number(valor) || 0));
    if (quantidade === 0) {
      delete this.quantidades[produtoId];
    } else {
      this.quantidades[produtoId] = quantidade;
    }
  }

  quantidadeDe(produtoId: number | undefined): number {
    return produtoId !== undefined ? (this.quantidades[produtoId] ?? 0) : 0;
  }

  get itensSelecionados(): ItemPayload[] {
    return Object.entries(this.quantidades).map(([produtoId, quantidade]) => ({
      produtoId: Number(produtoId),
      quantidade,
    }));
  }

  get totalItensSelecionados(): number {
    return this.itensSelecionados.length;
  }

  itensDaNota(nota: NotaFiscal): string {
    if (!nota.itens?.length) return '—';
    return nota.itens
      .map((item) => `${item.descricao || 'produto #' + item.produtoId} x${item.quantidade}`)
      .join(', ');
  }

  // Regra do negócio: só nota Aberta pode ser impressa.
  podeImprimir(nota: NotaFiscal): boolean {
    return nota.status === StatusNotaFiscal.Aberta;
  }

  salvar(): void {
    const itens = this.itensSelecionados;

    if (itens.length === 0) {
      this.mensagem = { tipo: 'erro', texto: 'Informe a quantidade de ao menos um produto.' };
      return;
    }

    // Validação client-side só pra feedback rápido — quem garante de
    // verdade é o Serviço de Estoque, no momento da impressão.
    const itemAcimaDoSaldo = itens.find((item) => {
      const produto = this.produtosDisponiveis.find((p) => p.id === item.produtoId);
      return produto !== undefined && item.quantidade > produto.saldo;
    });

    if (itemAcimaDoSaldo) {
      const produto = this.produtosDisponiveis.find((p) => p.id === itemAcimaDoSaldo.produtoId);
      this.mensagem = {
        tipo: 'erro',
        texto: `Quantidade de "${produto?.descricao}" maior que o saldo disponível (${produto?.saldo}).`,
      };
      return;
    }

    const payload: CriarNotaFiscalPayload = { itens };

    this.salvando = true;
    this.mensagem = null;

    this.notaFiscalService.criar(payload).subscribe({
      next: (nota) => {
        this.salvando = false;
        this.mensagem = {
          tipo: 'ok',
          texto: `Nota fiscal ${nota.numeroFormatado} criada como Aberta.`,
        };
        this.quantidades = {};
        this.carregarDados();
      },
      error: (err) => {
        this.salvando = false;
        this.mensagem = { tipo: 'erro', texto: this.extrairMensagemErro(err) };
      },
    });
  }

  imprimir(nota: NotaFiscal): void {
    if (!nota.id || !this.podeImprimir(nota) || this.imprimindoId !== null) return;

    this.imprimindoId = nota.id;
    this.mensagem = null;

    this.notaFiscalService.imprimir(nota.id).subscribe({
      next: (notaFechada) => {
        this.imprimindoId = null;
        this.mensagem = {
          tipo: 'ok',
          texto: `Nota ${notaFechada.numeroFormatado} impressa e fechada. Estoque atualizado.`,
        };
        this.carregarDados();
      },
      error: (err) => {
        this.imprimindoId = null;
        this.mensagem = { tipo: 'erro', texto: this.extrairMensagemErro(err) };
      },
    });
  }

  private extrairMensagemErro(err: { error?: unknown }): string {
    const corpo = err?.error as Partial<CorpoErroNegocio> | string | undefined;

    if (corpo && typeof corpo === 'object') {
      if (corpo.erro === 'saldo_insuficiente') {
        return `Saldo insuficiente para "${corpo.produto}": disponível ${corpo.disponivel}, solicitado ${corpo.solicitado}.`;
      }
      if (corpo.erro === 'estoque_indisponivel') {
        return corpo.mensagem ?? 'Serviço de estoque indisponível no momento. Tente novamente.';
      }
      if (corpo.erro === 'nota_nao_aberta') {
        return corpo.mensagem ?? 'Essa nota não está mais aberta.';
      }
      if (corpo.mensagem) {
        return corpo.mensagem;
      }
    }

    if (typeof corpo === 'string' && corpo.trim()) {
      return corpo;
    }

    return 'Erro ao processar a solicitação. Tente novamente.';
  }
}
