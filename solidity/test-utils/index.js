"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
exports.__esModule = true;
exports.deployContracts = void 0;
var hardhat_1 = require("hardhat");
var pure_1 = require("./pure");
function deployContracts(gravityId, validators, powers, powerThreshold, opts) {
    if (gravityId === void 0) { gravityId = "foo"; }
    return __awaiter(this, void 0, void 0, function () {
        var TestERC20, testERC20, Gravity, valAddresses, checkpoint, gravity, _a, _b, _c;
        return __generator(this, function (_d) {
            switch (_d.label) {
                case 0: 
                // enable automining for these tests
                return [4 /*yield*/, hardhat_1.ethers.provider.send("evm_setAutomine", [true])];
                case 1:
                    // enable automining for these tests
                    _d.sent();
                    return [4 /*yield*/, hardhat_1.ethers.getContractFactory("TestERC20A")];
                case 2:
                    TestERC20 = _d.sent();
                    return [4 /*yield*/, TestERC20.deploy()];
                case 3:
                    testERC20 = (_d.sent());
                    return [4 /*yield*/, hardhat_1.ethers.getContractFactory("Gravity")];
                case 4:
                    Gravity = _d.sent();
                    return [4 /*yield*/, (0, pure_1.getSignerAddresses)(validators)];
                case 5:
                    valAddresses = _d.sent();
                    checkpoint = (0, pure_1.makeCheckpoint)(valAddresses, powers, 0, 0, pure_1.ZeroAddress, gravityId);
                    _b = (_a = Gravity).deploy;
                    _c = [gravityId,
                        powerThreshold];
                    return [4 /*yield*/, (0, pure_1.getSignerAddresses)(validators)];
                case 6: return [4 /*yield*/, _b.apply(_a, _c.concat([_d.sent(), powers]))];
                case 7:
                    gravity = (_d.sent());
                    return [4 /*yield*/, gravity.deployed()];
                case 8:
                    _d.sent();
                    return [2 /*return*/, { gravity: gravity, testERC20: testERC20, checkpoint: checkpoint }];
            }
        });
    });
}
exports.deployContracts = deployContracts;
