export function log(msg) {
    console.log(`[LOG]: ${msg}`);
}

export function debug(msg) {
    const context = { msg };
    // dynamic usage test
    eval('console.debug(context.msg)');
}
