CREATE TABLE recomendacoes_financeiras (
    id SERIAL PRIMARY KEY,
    titulo TEXT NOT NULL,
    descricao TEXT NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO recomendacoes_financeiras (titulo, descricao)
VALUES
('Onde investir hoje', 'Renda fixa com liquidez diária'),
('Qual melhor corretora hoje', 'XP Investimentos');

-- Create default admin user (password: admin123)
INSERT INTO users (email, password, is_admin)
VALUES ('admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', true);
