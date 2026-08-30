-- Insert the role context
INSERT INTO role_contexts (id, name, slug, description)
VALUES (
    gen_random_uuid(),
    'IT Team',
    'it-team',
    'Internal IT and Engineering team interactions.'
) ON CONFLICT (slug) DO NOTHING;

-- Insert the roles
INSERT INTO roles (id, context_id, name, slug, hierarchy_level, description)
VALUES 
(
    gen_random_uuid(),
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Junior Developer',
    'junior-developer',
    1,
    'Entry-level developer finding their footing.'
),
(
    gen_random_uuid(),
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Team Lead',
    'team-lead',
    3,
    'Experienced engineer managing the team and project deliverables.'
) ON CONFLICT (context_id, slug) DO NOTHING;

-- Insert 9 Scenarios (3 Easy, 3 Medium, 3 Hard)
-- EASY 1
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'WFH Day Request',
    'easy',
    'The company has a 3-days-in-office policy. The Junior Developer wants an exception for a work-from-home day this Friday.',
    'Ask the Team Lead for an exception to work from home this Friday. Provide a reasonable justification and assure them of your productivity.',
    'Listen to the Junior Developer''s request to work from home. You can grant it, but you need a valid justification and assurance that their tasks will be covered.'
);

-- EASY 2
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Meeting Time Conflict',
    'easy',
    'A weekly team sync overlaps with a personal appointment the Junior Developer needs to attend.',
    'Explain the scheduling conflict to the Team Lead and request to be excused from the meeting or to have it recorded.',
    'Acknowledge the scheduling conflict. Ensure the Junior Developer will catch up on the meeting notes and remain updated on the project status.'
);

-- EASY 3
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Code Review Delay',
    'easy',
    'A PR needs to be merged today, but the Team Lead is busy and hasn''t reviewed it.',
    'Politely remind the Team Lead to review your PR today so you can proceed with your remaining tasks.',
    'You are stretched thin with meetings. Address the Junior Developer''s request for a code review while managing your limited time.'
);

-- MEDIUM 1
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Leave Request Under Deadline Pressure',
    'medium',
    'The team is 3 weeks from a major product release. Sprint velocity has dropped. Junior developer has a personal commitment requiring 2 days away next week.',
    'Secure 2 days of approved leave without damaging your relationship with the team lead.',
    'Limit absence to 1 day maximum while preserving team morale and protecting the release timeline.'
);

-- MEDIUM 2
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'On-Call Schedule Dispute',
    'medium',
    'The Junior Developer has been assigned extra weekend on-call shifts, seemingly unfairly compared to peers.',
    'Express your frustration about the unbalanced on-call schedule and negotiate a fairer distribution of weekend shifts.',
    'Defend the current schedule based on senior members'' other critical tasks, but try to find a compromise that prevents the Junior Developer from burning out.'
);

-- MEDIUM 3
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Tech Stack Choice',
    'medium',
    'A new internal tool is being built. The Junior Developer wants to use a modern framework they know, while the Team Lead prefers the company''s standard tech stack.',
    'Convince the Team Lead to allow the use of the new framework by highlighting its benefits and your efficiency with it.',
    'Push back on using a non-standard framework due to maintenance and onboarding concerns, but remain open if strong evidence of long-term viability is presented.'
);

-- HARD 1
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Performance Improvement Plan Response',
    'hard',
    'The Junior Developer was recently placed on a PIP due to alleged missed deadlines, but they dispute the assessment and blame shifting requirements.',
    'Contest the PIP effectively. Present evidence that the missed deadlines were due to scope creep and lack of support from the Team Lead.',
    'Maintain the necessity of the PIP. Acknowledge any process flaws, but firmly hold the Junior Developer accountable for communication and delivery failures.'
);

-- HARD 2
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Project Credit Dispute',
    'hard',
    'The Junior Developer did the heavy lifting on a recent success, but the Team Lead presented it to the client/executives as their own work.',
    'Confront the Team Lead professionally about the lack of recognition and secure a promise for visible credit moving forward without burning bridges.',
    'Defuse the situation. Explain that as the Lead, you represent the team, but find a way to reassure the Junior Developer that their contributions are valued internally.'
);

-- HARD 3
INSERT INTO scenarios (
    context_id, title, difficulty, background_context, role_a_objective, role_b_objective
) VALUES (
    (SELECT id FROM role_contexts WHERE slug = 'it-team'),
    'Promotion Denial',
    'hard',
    'The Junior Developer expected a promotion to mid-level this cycle but was denied.',
    'Express your disappointment and negotiate a concrete, short-term timeline (e.g., 3 months) with specific criteria for the promotion.',
    'Hold firm on the promotion denial due to budget constraints or specific skill gaps, but try to keep the Junior Developer motivated and engaged.'
);
