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
  
  quantidades: Record<number, number> = {};   // record direciona qual produto com qual quantidade

  carregando = false;
  salvando = false;
  imprimindoId: number | null = null; // guarda qual nota está sendo impressa
  mensagem: { tipo: 'ok' | 'erro'; texto: string } | null = null;

  private notaFiscalService = inject(NotaFiscalService); // injeção de dependências
  private produtoService = inject(ProdutoService);

  ngOnInit(): void { // quando inicializa o componente carrega os dados
    this.carregarDados();
  }

  carregarDados(): void {
    this.carregando = true; // informa que está carregando
    this.produtoService.listar().subscribe({ // chama o service para escrever os produtos disponiveis
      next: (produtos) => {
        this.produtosDisponiveis = produtos;
        this.carregando = false;
      },
      error: () => { // tratamento de erro
        this.carregando = false;
        this.mensagem = { tipo: 'erro', texto: 'Não foi possível carregar os produtos.' };
      },
    });

    this.notaFiscalService.listar().subscribe({ // chama service para escrever as notas
      next: (notas) => (this.notas = notas),
      error: () => { // tratamento de erro
        this.mensagem = { tipo: 'erro', texto: 'Não foi possível carregar as notas fiscais.' };
      },
    });
  }

  atualizarQuantidade(produtoId: number | undefined, valor: string): void {
    if (produtoId === undefined) return;     // precisa de um produto para alterar quantidade

    const quantidade = Math.max(0, Math.floor(Number(valor) || 0));  // garante que não fique negativo e muda e não saia numero quebrado
    if (quantidade === 0) { // se alterar para zero deleta a quantidade
      delete this.quantidades[produtoId];
    } else { // se tiver algum outro número ele recebe o valor colocado
      this.quantidades[produtoId] = quantidade;
    }
  }

  quantidadeDe(produtoId: number | undefined): number {
    return produtoId !== undefined ? (this.quantidades[produtoId] ?? 0) : 0;
  } // serve para retornar a quantidade do produto

  get itensSelecionados(): ItemPayload[] {
    return Object.entries(this.quantidades).map(([produtoId, quantidade]) => ({
      produtoId: Number(produtoId),  // Number() pois objeto retornado vem como string
      quantidade,
    }));
  }

  get totalItensSelecionados(): number {
    return this.itensSelecionados.length; // pega quantidade de itens selecionado apenas
  }

  itensDaNota(nota: NotaFiscal): string { // transforma os itens json em uma string para exibir na tela
    if (!nota.itens?.length) return '—';
    return nota.itens
      .map((item) => `${item.descricao || 'produto #' + item.produtoId} x${item.quantidade}`)
      .join(', ');
  }

  // Regra do negócio: só nota Aberta pode ser impressa.
  podeImprimir(nota: NotaFiscal): boolean {
    return nota.status === StatusNotaFiscal.Aberta; // verificar se está aberta
  }

  salvar(): void {
    const itens = this.itensSelecionados; // recebe os itens selecionados

    if (itens.length === 0) { // se não existir itens informa erro
      this.mensagem = { tipo: 'erro', texto: 'Informe a quantidade de ao menos um produto.' };
      return;
    }

    // Validação client-side só pra feedback rápido — quem garante de
    // verdade é o Serviço de Estoque, no momento da impressão.
    const itemAcimaDoSaldo = itens.find((item) => {
      const produto = this.produtosDisponiveis.find((p) => p.id === item.produtoId);
      return produto !== undefined && item.quantidade > produto.saldo;
    });

    if (itemAcimaDoSaldo) { // se condição verdadeira informa erro de quantidade desejada maior que saldo
      const produto = this.produtosDisponiveis.find((p) => p.id === itemAcimaDoSaldo.produtoId);
      this.mensagem = {
        tipo: 'erro',
        texto: `Quantidade de "${produto?.descricao}" maior que o saldo disponível (${produto?.saldo}).`,
      };
      return;
    }

    const payload: CriarNotaFiscalPayload = { itens }; // criando o DTO esperado pelo backend

    this.salvando = true; // indica que está salvando
    this.mensagem = null;

    this.notaFiscalService.criar(payload).subscribe({ // chama o service para criar
      next: (nota) => { // recebe nota de volta criada
        this.salvando = false;
        this.mensagem = {
          tipo: 'ok',
          texto: `Nota fiscal ${nota.numeroFormatado} criada como Aberta.`, // informa o numero da nota
        };
        this.quantidades = {}; // limpa os produtos selecionados
        this.carregarDados(); // carrega novamente a tela
      },
      error: (err) => { // tratamento de erro
        this.salvando = false;
        this.mensagem = { tipo: 'erro', texto: this.extrairMensagemErro(err) };
      },
    });
  }

  imprimir(nota: NotaFiscal): void { // verifica ID , se pode imprimir e se já está imprimindo alguma nota
    if (!nota.id || !this.podeImprimir(nota) || this.imprimindoId !== null) return;

    this.imprimindoId = nota.id;  // declara que imprimindoID recebe o id da nota atual
    this.mensagem = null; // zera mensagem

    this.notaFiscalService.imprimir(nota.id).subscribe({ // chama service para imprimir
      next: (notaFechada) => { // recebe nota fechada
        this.imprimindoId = null; // limpa o imprimindoID
        this.mensagem = { // retorna mensagem de sucesso
          tipo: 'ok',
          texto: `Nota ${notaFechada.numeroFormatado} impressa e fechada. Estoque atualizado.`,
        }; // carrega tela novamente
        this.carregarDados();
      },
      error: (err) => { // tratamento de erro
        this.imprimindoId = null;
        this.mensagem = { tipo: 'erro', texto: this.extrairMensagemErro(err) };
      },
    });
  }

  private extrairMensagemErro(err: { error?: unknown }): string {
    const corpo = err?.error as Partial<CorpoErroNegocio> | string | undefined;
    // função de extrair mensagem usada no tratamento de erro das últimas funções
    if (corpo && typeof corpo === 'object') {
      if (corpo.erro === 'saldo_insuficiente') { // verifica qual o erro
        return `Saldo insuficiente para "${corpo.produto}": disponível ${corpo.disponivel}, solicitado ${corpo.solicitado}.`;
      } // retorna a mensagem adequada
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
