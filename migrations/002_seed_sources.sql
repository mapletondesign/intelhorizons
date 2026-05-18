INSERT INTO sources (name, type, url, fetch_interval_mins) VALUES

-- RSS: Dedicated AI publications
('The Decoder',             'rss', 'https://the-decoder.com/feed/',                                                                    30),
('VentureBeat AI',          'rss', 'https://venturebeat.com/category/ai/feed/',                                                        30),
('MIT Technology Review AI','rss', 'https://www.technologyreview.com/topic/artificial-intelligence/feed',                               60),
('TechCrunch AI',           'rss', 'https://techcrunch.com/category/artificial-intelligence/feed/',                                    30),
('The Verge AI',            'rss', 'https://www.theverge.com/rss/ai-artificial-intelligence/index.xml',                                30),
('Ars Technica',            'rss', 'https://feeds.arstechnica.com/arstechnica/index',                                                  60),
('Wired AI',                'rss', 'https://www.wired.com/feed/tag/ai/latest/rss',                                                     60),
('IEEE Spectrum AI',        'rss', 'https://spectrum.ieee.org/feeds/topic/artificial-intelligence.rss',                                60),
('Analytics Vidhya',        'rss', 'https://www.analyticsvidhya.com/feed/',                                                            60),
('Bloomberg Technology',    'rss', 'https://feeds.bloomberg.com/technology/news.rss',                                                  60),
('Reuters Technology',      'rss', 'https://feeds.reuters.com/reuters/technologyNews',                                                 60),

-- RSS: Newsletters / Substacks
('The Batch',               'rss', 'https://www.deeplearning.ai/the-batch/feed/',                                                      360),
('Import AI',               'rss', 'https://importai.substack.com/feed',                                                               360),
('Ahead of AI',             'rss', 'https://magazine.sebastianraschka.com/feed',                                                       360),

-- RSS: Lab blogs
('Hugging Face Blog',       'rss', 'https://huggingface.co/blog/feed.xml',                                                             60),
('Papers With Code',        'rss', 'https://paperswithcode.com/rss',                                                                   60),

-- ArXiv: cs.AI and cs.LG daily
('ArXiv cs.AI',             'arxiv', 'https://export.arxiv.org/api/query?search_query=cat:cs.AI&max_results=30&sortBy=submittedDate&sortOrder=descending', 120),
('ArXiv cs.LG',             'arxiv', 'https://export.arxiv.org/api/query?search_query=cat:cs.LG&max_results=30&sortBy=submittedDate&sortOrder=descending', 120),

-- HN: AI-filtered stories
('Hacker News AI',          'hn',    'https://hn.algolia.com/api/v1/search_by_date?tags=story&query=AI+LLM+machine+learning&hitsPerPage=30', 30),

-- Reddit: read-only JSON API
('Reddit r/MachineLearning','reddit','https://www.reddit.com/r/MachineLearning/new.json?limit=25',  30),
('Reddit r/artificial',     'reddit','https://www.reddit.com/r/artificial/new.json?limit=25',       30),
('Reddit r/LocalLLaMA',     'reddit','https://www.reddit.com/r/LocalLLaMA/new.json?limit=25',       30),

-- Twitter/X: bearer-token required; inactive until configured
('Reddit r/Singularity',    'reddit','https://www.reddit.com/r/singularity/new.json?limit=25',      30)

ON CONFLICT (name) DO NOTHING;

INSERT INTO sources (name, type, url, config, active, fetch_interval_mins)
VALUES (
    'X — AI Researchers',
    'twitter',
    'https://api.twitter.com/2/users/by',
    '{"handles":"karpathy,sama,ylecun,fchollet,demishassabis,goodside,alexalbert__,GaryMarcus,drfeifei,emollick"}',
    false,
    60
) ON CONFLICT (name) DO NOTHING;
