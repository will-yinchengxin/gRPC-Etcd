Go中命名为internal的package，只有该package的父级package才可以访问该package的内容。
> 例如，一个包的路径.../a/b/c/internal/d/e/f只能被.../a/b/c的代码层级包引入，不能被.../a/b/g或其他的任意目录引用；

````

````

>内部包的规范约定：导入路径包含internal关键字的包，只允许internal的父级目录及父级目录的子包导入，其它包无法导入。
 

[官方文档](https://golang.org/s/go14internal) https://golang.org/s/go14internal

两点注意：
- 只有直接父级package可以访问，其他的都不行，再往上的祖先package也不行
- 父级package也只能访问internal package使用大写暴露出的内容，小写的不行
