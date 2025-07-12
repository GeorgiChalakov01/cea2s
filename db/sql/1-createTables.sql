BEGIN;

CREATE TABLE part1_questions (
	id SERIAL PRIMARY KEY,
	question_text TEXT NOT NULL,
	audio_filename VARCHAR(255) NOT NULL
);

INSERT INTO part1_questions (question_text, audio_filename) VALUES
('What is your name?', 'question-1.mp3'),
('Where are you from?', 'question-2.mp3'),
('Do you like doing any sports?', 'question-3.mp3'),
('Do you like traveling?', 'question-4.mp3'),
('Do you have any hobbies?', 'question-5.mp3'),
('Do you like speaking face to face with people, or do you prefer texting them?', 'question-6.mp3'),
('In your opinion what is the impact on technology on the job market?', 'question-7.mp3'),
('Do you prefer driving or walking?', 'question-8.mp3'),
('Do you prefer working from home or from the office?', 'question-9.mp3'),
('Do you have any pets and would you like to get some?', 'question-10.mp3');

COMMIT;
