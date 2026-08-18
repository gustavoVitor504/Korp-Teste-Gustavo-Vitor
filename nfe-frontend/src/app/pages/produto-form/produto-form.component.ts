import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ProdutoService } from '../../services/produto.service';
import { Produto } from '../../models/produto.model';

@Component({
  selector: 'app-produto-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './produto-form.component.html',
  styleUrl: './produto-form.component.css',
})
export class ProdutoFormComponent implements OnInit {
  produtos: Produto[] = [];
  carregando = false;
  salvando = false;
  mensagem: { tipo: 'ok' | 'erro'; texto: string } | null = null;

  private fb = inject(FormBuilder);
  private produtoService = inject(ProdutoService);

  form = this.fb.group({
    codigo: ['', [Validators.required, Validators.minLength(1)]],
    descricao: ['', [Validators.required, Validators.minLength(2)]],
    saldo: [0, [Validators.required, Validators.min(0)]],
  });

  ngOnInit(): void {
    this.carregarProdutos();
  }

  carregarProdutos(): void {
    this.carregando = true;
    this.produtoService.listar().subscribe({
      next: (produtos) => {
        this.produtos = produtos;
        this.carregando = false;
      },
      error: () => {
        this.carregando = false;
        this.mensagem = { tipo: 'erro', texto: 'Não foi possível carregar os produtos.' };
      },
    });
  }

  salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const produto: Produto = {
      codigo: this.form.value.codigo!.trim(),
      descricao: this.form.value.descricao!.trim(),
      saldo: this.form.value.saldo!,
    };

    this.salvando = true;
    this.mensagem = null;

    this.produtoService.criar(produto).subscribe({
      next: () => {
        this.salvando = false;
        this.mensagem = { tipo: 'ok', texto: `Produto "${produto.descricao}" cadastrado.` };
        this.form.reset({ codigo: '', descricao: '', saldo: 0 });
        this.carregarProdutos();
      },
      error: (err) => {
        this.salvando = false;
        const texto = err?.error?.message ?? 'Erro ao cadastrar produto. Verifique os dados.';
        this.mensagem = { tipo: 'erro', texto };
      },
    });
  }
}
