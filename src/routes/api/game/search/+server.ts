import type { RequestHandler } from '@sveltejs/kit';
import bggXmlApiClient from 'bgg-xml-api-client';

export const GET: RequestHandler = async ({ url }) => {
	const query = url.searchParams.get('q');
	if (!query) return new Response('Missing query', { status: 400 });

	const queryResults = await bggXmlApiClient.getBggSearch({
		query,
		type: 'boardgame'
	});

	const ids = queryResults.item.map((item) => item.id);
	const games = await bggXmlApiClient.getBggThing({
		id: ids,
		type: 'boardgame'
	});

	return new Response(JSON.stringify({ query, games }), { status: 200 });
};
