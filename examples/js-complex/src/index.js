import { Button } from './components/Button';
import { add } from './utils/math';
import { successIcon } from './components/Icon';
import { log } from './utils/logger';

function main() {
    log('App started');
    const b = Button();
    const result = add(10, 20);
    console.log(result, successIcon);
}

main();
