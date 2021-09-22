import {isEmpty} from "../utils/utils";
import request from "../common/request";

export function hasPermission(owner) {
    let userJsonStr = sessionStorage.getItem('user');
    if (isEmpty(userJsonStr)) {
        return false;
    }
    let user = JSON.parse(userJsonStr);
    if (user['type'] === 'admin') {
        return true;
    }

    return user['id'] === owner;
}

export function isAdmin() {
    let userJsonStr = sessionStorage.getItem('user');
    if (isEmpty(userJsonStr)) {
        return false;
    }
    let user = JSON.parse(userJsonStr);
    return user['type'] === 'admin';
}

export async function allowview(name) {
    let ntcfgJsonStr = sessionStorage.getItem('ntcfg');
    if (isEmpty(ntcfgJsonStr)) {
        let result = await request.get('/showcfg');
        console.log(result['code'])
        if (result['code'] === 200) {
            sessionStorage.setItem('ntcfg', JSON.stringify(result['data']))
            return result['data'][name]
        }
        return false
    }
    let ntcfg = JSON.parse(ntcfgJsonStr);
    return ntcfg[name]
}

export function getCurrentUser() {
    let userJsonStr = sessionStorage.getItem('user');
    if (isEmpty(userJsonStr)) {
        return {};
    }

    return JSON.parse(userJsonStr);
}