CREATE TABLE recomendacoes_financeiras (
    id SERIAL PRIMARY KEY,
    titulo TEXT NOT NULL,
    descricao TEXT NOT NULL
);

INSERT INTO recomendacoes_financeiras (titulo, descricao)
VALUES 
('Onde investir hoje', 'Renda fixa com liquidez diária'),
('Qual melhor corretora hoje', 'XP Investimentos');
