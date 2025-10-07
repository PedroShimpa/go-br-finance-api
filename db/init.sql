CREATE TABLE IF NOT EXISTS recomendacoes_financeiras (
    id SERIAL PRIMARY KEY,
    titulo TEXT NOT NULL,
    descricao TEXT NOT NULL
);

-- Add new columns if they don't exist
ALTER TABLE recomendacoes_financeiras ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE recomendacoes_financeiras ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Insert initial data if table is empty
INSERT INTO recomendacoes_financeiras (titulo, descricao)
SELECT 'Onde investir hoje', 'Renda fixa com liquidez diária'
WHERE NOT EXISTS (SELECT 1 FROM recomendacoes_financeiras WHERE titulo = 'Onde investir hoje');

INSERT INTO recomendacoes_financeiras (titulo, descricao)
SELECT 'Qual melhor corretora hoje', 'XP Investimentos'
WHERE NOT EXISTS (SELECT 1 FROM recomendacoes_financeiras WHERE titulo = 'Qual melhor corretora hoje');
