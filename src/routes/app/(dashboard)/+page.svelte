<script lang="ts">
	import { enhance } from '$app/forms';

	import type { PageData } from './$types';

	export let data: PageData;
</script>

<h1>Profile</h1>
<pre>{JSON.stringify(data, null, 2)}</pre>
<form method="post" action="?/logout" use:enhance>
	<input type="submit" value="Sign out" />
</form>
