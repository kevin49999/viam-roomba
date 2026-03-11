<script lang="ts">
	import { MotionTools } from '@viamrobotics/motion-tools';
	import { DEFAULT_PART_ID } from '$lib/consts';
	import { Struct } from '@viamrobotics/sdk';
</script>


<div class="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
    <MotionTools
        partID={DEFAULT_PART_ID}
        localConfigProps={{
            current: new Struct(),
            setLocalPartConfig: () => {},
            isDirty: false,
            componentToFragId: {}
        }}
    />
</div>
