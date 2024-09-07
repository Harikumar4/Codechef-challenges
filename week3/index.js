import express from 'express';
import dotenv from 'dotenv';
import { GoogleGenerativeAI } from '@google/generative-ai';

// Initialize environment variables
dotenv.config({ path: 'hello.env' });

// Initialize Express app
const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(express.json());
app.use(express.static('public'));  // Serve static files from the 'public' directory

// Serve the index.html file when the root route is accessed
app.get('/', (req, res) => {
    res.sendFile(__dirname + '/public/index.html');
});

// Initialize Google Generative AI Client
const API_KEY = process.env.GOOGLE_API_KEY;
const genAI = new GoogleGenerativeAI(API_KEY);
const model = genAI.getGenerativeModel({ model: "gemini-1.5-flash" });

// Generate HTML based on template and content
app.post('/generate-html', async (req, res) => {
    const { template, content } = req.body;
    try {
        const prompt = `Generate HTML using this template: ${template} with this content: ${content}`;
        const result = await model.generateContent(prompt);
        res.json({ generatedHtml: result.response.text() });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Modify existing HTML based on user instructions
app.post('/modify-html', async (req, res) => {
    const { html, instructions } = req.body;
    try {
        const prompt = `Modify this HTML: ${html} based on these instructions: ${instructions}`;
        const result = await model.generateContent(prompt);

        // Log the entire result for debugging
        console.log('API Response:', result);

        // Check if result.response and result.response.text() are valid
        if (result && result.response && typeof result.response.text === 'function') {
            res.json({ modifiedHtml: result.response.text() });
        } else {
            res.status(500).json({ error: 'Unexpected response format from the API' });
        }
    } catch (error) {
        // Log detailed error
        console.error('Error during API call:', error);
        res.status(500).json({ error: error.message || 'Internal Server Error' });
    }
})

// Start the server
app.listen(PORT, () => {
    console.log(`Server running at http://localhost:${PORT}`);
});
